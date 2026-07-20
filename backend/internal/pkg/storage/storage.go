// Package storage 负责图片二进制的本地落盘与对外访问 URL 的拼接。
//
// 设计目标是「薄、可替换」：service 层只依赖这里导出的 Saver，
// 后续若要换成对象存储（S3/OSS/COS 等），只需提供一个等价的实现即可，
// 无需触碰业务逻辑。
//
// 目录结构约定：<root>/<YYYY>/<MM>/<sha256>.<ext>
// 这种按年月分桶 + hash 文件名的方式既能避免单目录文件过多，
// 又天然支持秒传（同一张图片必然落到同一路径）。
package storage

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
)

// Saver 是本地文件存储器。
//
// rootDir 是文件实际落盘的根目录（相对路径相对于进程 cwd，建议生产用绝对路径）。
// publicBaseURL 用于拼接对外访问 URL，空字符串表示走相对路径 "/imgs/..."。
type Saver struct {
	rootDir       string
	publicBaseURL string
}

// 相对 URL 前缀。Nginx 反代本地图片目录时挂在该路径下，配置侧无需另外配。
const relURLPrefix = "/imgs"

// fileMode 是落盘图片文件的权限：0644（属主可读写，group/other 只读）。
// os.CreateTemp 创建的临时文件默认 0600，而 os.Rename 只改名不改权限，故最终落盘文件仍是 0600；
// 生产环境 nginx(www) 作为 other 无法读取会直接 403。这里显式 Chmod 为 0644 保证 web server 跨用户可读。
const fileMode os.FileMode = 0o644

// NewSaver 基于配置构造 Saver，并提前 MkdirAll 出 rootDir 以便快速暴露权限问题。
func NewSaver(cfg config.StorageConfig) (*Saver, error) {
	root := strings.TrimSpace(cfg.RootDir)
	if root == "" {
		return nil, errors.New("storage.root_dir 不能为空")
	}
	// 拒绝把存储根目录配成「会暴露进程工作目录」的危险路径：root 是 cwd 本身或其祖先
	// （含 "." / ".." / "/" 等）时，/imgs 静态服务会把 config.yaml / irisImg.db / 源码
	// 未认证暴露出去。即便 serveImages 有扩展名白名单兜底，这里也在启动期 fail-fast，
	// 把用户误配扼杀在启动阶段。
	if err := guardRootDir(root); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("创建存储根目录失败: %w", err)
	}

	// public_base_url 允许尾部带或不带斜杠，统一去掉，拼接时由我们补。
	base := strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	// 容错裸域名：非空且无协议前缀时自动补 https://。
	// 否则前端会把无协议的值当相对路径解析（如 "img.example.com/imgs" -> "/img.example.com/imgs/..."），
	// 导致图片 src 错乱。已带 http:// / https:// 的原样保留。
	if base != "" && !strings.Contains(base, "://") {
		base = "https://" + base
	}

	return &Saver{rootDir: root, publicBaseURL: base}, nil
}

// guardRootDir 拒绝 storage.root_dir 指向后端工作目录本身或其祖先的危险配置。
//
// 一旦 root 是 cwd 或 cwd 的祖先（如 root_dir 配成 "." / ".." / "/" / 工作目录父级），
// /imgs 静态服务即会把工作目录下的 config.yaml / irisImg.db / 源码暴露给未认证访客。
// serveImages 的扩展名白名单是兜底，这里在启动期 fail-fast 把误配扼杀得更早、错误更明确。
//
// 判定依据：filepath.Rel(absRoot, absCwd) 得到「cwd 相对 root 的路径」，
// 不以 ".." 开头（含 "."）说明 cwd 落在 root 之内或等于 root -> root 是 cwd 本身/祖先 -> 拒绝。
// 取不到 cwd 或无法定位时保守放行（扩展名白名单仍是兜底）。
func guardRootDir(root string) error {
	absRoot, err := filepath.Abs(filepath.Clean(root))
	if err != nil {
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	absCwd, err := filepath.Abs(filepath.Clean(cwd))
	if err != nil {
		return nil
	}
	rel, err := filepath.Rel(absRoot, absCwd)
	if err != nil {
		// 跨盘符等无法计算相对路径的情况，保守放行，交由扩展名白名单兜底。
		return nil
	}
	sep := string(filepath.Separator)
	outside := rel == ".." || strings.HasPrefix(rel, ".."+sep)
	if !outside {
		return fmt.Errorf("storage.root_dir（%s）不能是后端工作目录（%s）本身或其祖先："+
			"/imgs 静态服务会未认证暴露工作目录下的 config.yaml / 数据库 / 源码；"+
			"请改为工作目录之外的专用目录（如 /var/lib/irisImg/imgs）或工作目录之下的独立子目录（如 data/imgs）", absRoot, absCwd)
	}
	return nil
}

// RelPath 根据时间与文件名计算用于落盘的相对路径（"YYYY/MM/<hash>.<ext>"）。
// 单独导出便于上层在不真正写盘的情况下预判落盘位置（如秒传分支）。
//
// ext 不含点号，会被强制小写化；为空时退化为 "bin"。
func RelPath(hash, ext string, t time.Time) string {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	if ext == "" {
		ext = "bin"
	}
	// path.Join 强制使用正斜杠，便于做 URL 拼接，不会污染 Windows 文件系统
	// （写盘时再用 filepath.Join 转回平台分隔符）。
	return path.Join(t.Format("2006"), t.Format("01"), hash+"."+ext)
}

// Save 把内容写到 <root>/<YYYY>/<MM>/<hash>.<ext>，返回相对路径（始终用正斜杠）。
//
// 如果目标文件已经存在（同 hash 二次写入），直接复用不报错，保证幂等。
// 写入流程是「同目录临时文件 + Rename」，避免半写状态污染目录。
func (s *Saver) Save(content []byte, hash, ext string, t time.Time) (string, error) {
	rel := RelPath(hash, ext, t)
	abs := filepath.Join(s.rootDir, filepath.FromSlash(rel))

	// 同 hash 已经写过 → 直接复用。
	if info, err := os.Stat(abs); err == nil && !info.IsDir() {
		return rel, nil
	}

	// 创建年月目录。
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 临时文件 + Rename，保证原子可见。
	tmp, err := os.CreateTemp(filepath.Dir(abs), ".upload-*")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	tmpPath := tmp.Name()
	// 出错时清理临时文件（成功路径下 tmp 已被 rename，删除会无害失败）。
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("关闭临时文件失败: %w", err)
	}

	if err := os.Rename(tmpPath, abs); err != nil {
		// Rename 在 Windows 下偶发目标已存在时失败：兜底再 Stat 一次，
		// 命中说明已被并发上传写过，按秒传处理。
		if info, statErr := os.Stat(abs); statErr == nil && !info.IsDir() {
			// 旧版本落盘为 0600，此处幂等纠正为 0644，避免历史文件被 nginx 跨用户读取时 403。
			_ = os.Chmod(abs, fileMode)
			return rel, nil
		}
		return "", fmt.Errorf("重命名落盘文件失败: %w", err)
	}

	// os.CreateTemp 以 0600 创建临时文件，os.Rename 只改名不改权限，故最终落盘文件仍是 0600。
	// 生产环境 nginx(www) 作为 other 无法读取会 403，这里显式改为 0644 保证 web server 跨用户可读。
	if err := os.Chmod(abs, fileMode); err != nil {
		return "", fmt.Errorf("设置文件权限失败: %w", err)
	}

	return rel, nil
}

// Delete 删除相对路径对应的物理文件，文件不存在视为已删除（幂等）。
// rel 采用正斜杠（与 Save 返回值一致），内部转回平台分隔符后拼接到 rootDir 之下。
func (s *Saver) Delete(rel string) error {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return nil
	}
	abs := filepath.Join(s.rootDir, filepath.FromSlash(rel))
	if err := os.Remove(abs); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// PublicURL 把相对路径拼接成对外访问 URL：
//   - publicBaseURL 为空 → "/imgs/<rel>"；
//   - 非空 → "<base>/<rel>"。
func (s *Saver) PublicURL(rel string) string {
	rel = strings.TrimLeft(rel, "/")
	if s.publicBaseURL == "" {
		return relURLPrefix + "/" + rel
	}
	return s.publicBaseURL + "/" + rel
}

// RootDir 返回存储根目录，供测试与诊断使用。
func (s *Saver) RootDir() string {
	return s.rootDir
}
