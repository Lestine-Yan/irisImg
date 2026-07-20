package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Lestine-Yan/irisImg/backend/config"
)

// fixed 是一个固定时间，便于按年月断言路径。
var fixed = time.Date(2026, 6, 23, 15, 30, 0, 0, time.UTC)

func newSaver(t *testing.T, baseURL string) *Saver {
	t.Helper()
	root := filepath.Join(t.TempDir(), "imgs")
	s, err := NewSaver(config.StorageConfig{RootDir: root, PublicBaseURL: baseURL})
	if err != nil {
		t.Fatalf("new saver: %v", err)
	}
	return s
}

func TestRelPath(t *testing.T) {
	got := RelPath("abc", "PNG", fixed)
	want := "2026/06/abc.png"
	if got != want {
		t.Fatalf("rel path = %q, want %q", got, want)
	}

	// 空扩展名兜底
	if got := RelPath("abc", "", fixed); got != "2026/06/abc.bin" {
		t.Fatalf("empty ext fallback = %q", got)
	}

	// 允许带点的扩展名
	if got := RelPath("abc", ".Jpg", fixed); got != "2026/06/abc.jpg" {
		t.Fatalf("dotted ext = %q", got)
	}
}

func TestSaver_Save_WritesFile(t *testing.T) {
	s := newSaver(t, "")
	rel, err := s.Save([]byte("hello"), "abc", "png", fixed)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if rel != "2026/06/abc.png" {
		t.Fatalf("rel = %q", rel)
	}

	abs := filepath.Join(s.RootDir(), "2026", "06", "abc.png")
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("content = %q", string(data))
	}
}

func TestSaver_Save_IdempotentOnSecondWrite(t *testing.T) {
	s := newSaver(t, "")

	rel, err := s.Save([]byte("v1"), "abc", "png", fixed)
	if err != nil {
		t.Fatalf("first save: %v", err)
	}

	// 二次写：即使内容不同，路径已存在则直接复用，不报错也不覆盖。
	// 这模拟「同 hash 已存在」的秒传分支兜底；service 上层会先查库，
	// 真实路径不会走到这一步，但 Saver 自己也得保证幂等。
	rel2, err := s.Save([]byte("ignored"), "abc", "png", fixed)
	if err != nil {
		t.Fatalf("second save: %v", err)
	}
	if rel != rel2 {
		t.Fatalf("rel mismatch: %q vs %q", rel, rel2)
	}

	abs := filepath.Join(s.RootDir(), "2026", "06", "abc.png")
	data, _ := os.ReadFile(abs)
	if string(data) != "v1" {
		t.Fatalf("file overwritten unexpectedly: %q", string(data))
	}
}

func TestSaver_PublicURL(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		rel     string
		want    string
	}{
		{"empty base url", "", "2026/06/abc.png", "/imgs/2026/06/abc.png"},
		{"absolute base", "https://img.example.com", "2026/06/abc.png", "https://img.example.com/2026/06/abc.png"},
		{"trim trailing slash", "https://img.example.com/", "2026/06/abc.png", "https://img.example.com/2026/06/abc.png"},
		{"leading slash on rel", "https://img.example.com", "/2026/06/abc.png", "https://img.example.com/2026/06/abc.png"},
		{"bare domain gets https", "img.example.com", "2026/06/abc.png", "https://img.example.com/2026/06/abc.png"},
		{"bare domain with path gets https", "img.example.com/imgs", "2026/06/abc.png", "https://img.example.com/imgs/2026/06/abc.png"},
		{"http scheme preserved", "http://img.example.com", "2026/06/abc.png", "http://img.example.com/2026/06/abc.png"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := newSaver(t, c.baseURL)
			if got := s.PublicURL(c.rel); got != c.want {
				t.Fatalf("url = %q, want %q", got, c.want)
			}
		})
	}
}

func TestNewSaver_EmptyRootDir(t *testing.T) {
	_, err := NewSaver(config.StorageConfig{RootDir: ""})
	if err == nil || !strings.Contains(err.Error(), "root_dir") {
		t.Fatalf("expected root_dir error, got %v", err)
	}
}

// TestNewSaver_RejectsDangerousRoot 验证 NewSaver 在启动期 fail-fast 拒绝把
// storage.root_dir 配成后端工作目录本身或其祖先（含 "." / ".."）。
// 这类配置会让 /imgs 静态服务把 config.yaml / irisImg.db / 源码暴露给未认证访客。
func TestNewSaver_RejectsDangerousRoot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	cases := []struct {
		name string
		root string
	}{
		{"cwd itself", cwd},
		{"dot equals cwd", "."},
		{"parent of cwd", ".."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := NewSaver(config.StorageConfig{RootDir: c.root})
			if err == nil {
				t.Fatalf("expected dangerous root %q to be rejected", c.root)
			}
			if !strings.Contains(err.Error(), "工作目录") {
				t.Fatalf("error should mention 工作目录, got %v", err)
			}
		})
	}
}

// TestNewSaver_AcceptsSafeRoot 验证「cwd 之下的独立子目录」（默认 data/imgs 形态）
// 与「cwd 之外的专用目录」均放行--这是安全且常见的部署形态，不应被 fail-fast 误伤。
func TestNewSaver_AcceptsSafeRoot(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// cwd 之下的独立子目录：默认 config.yaml 的 data/imgs 即此形态。
	subDir := filepath.Join(cwd, "testdata_safe_imgs")
	t.Cleanup(func() { _ = os.RemoveAll(subDir) })
	if _, err := NewSaver(config.StorageConfig{RootDir: subDir}); err != nil {
		t.Fatalf("subdir-under-cwd should be accepted, got %v", err)
	}

	// cwd 之外的临时目录：生产 /var/lib/irisImg/imgs 形态。
	external := filepath.Join(t.TempDir(), "imgs")
	if _, err := NewSaver(config.StorageConfig{RootDir: external}); err != nil {
		t.Fatalf("external dir should be accepted, got %v", err)
	}
}

// TestNewSaver_NormalizesBaseURL 验证裸域名 public_base_url 会被自动补 https://，
// 已带 http(s):// 的原样保留，空仍为空。防止前端把无协议值当相对路径解析导致 src 错乱。
func TestNewSaver_NormalizesBaseURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string // 期望的内部 base；空表示走相对路径 /imgs
	}{
		{"empty stays empty", "", ""},
		{"bare domain gets https", "img.example.com", "https://img.example.com"},
		{"bare domain with path", "img.example.com/imgs", "https://img.example.com/imgs"},
		{"trailing slash trimmed then normalized", "img.example.com/", "https://img.example.com"},
		{"https preserved", "https://img.example.com", "https://img.example.com"},
		{"http preserved", "http://img.example.com", "http://img.example.com"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := NewSaver(config.StorageConfig{RootDir: filepath.Join(t.TempDir(), "imgs"), PublicBaseURL: c.in})
			if err != nil {
				t.Fatalf("new saver: %v", err)
			}
			// 用 PublicURL 间接验证内部 publicBaseURL：空 -> "/imgs/x"，非空 -> "<base>/x"。
			got := s.PublicURL("x")
			want := c.want + "/x"
			if c.want == "" {
				want = "/imgs/x"
			}
			if got != want {
				t.Fatalf("PublicURL(x) = %q, want %q", got, want)
			}
		})
	}
}

// TestSaver_Save_FileMode 验证落盘文件权限为 0644（属主 rw、group/other r）。
// os.CreateTemp 默认 0600，若不显式 Chmod，生产环境 nginx(www) 跨用户读取会 403。
func TestSaver_Save_FileMode(t *testing.T) {
	// Windows 不支持 Unix 权限位，os.Stat 对可读写文件恒返回 0666，无法验证 0644。
	// 生产为 Linux，os.Chmod(0644) 在那里才真正生效，故本用例仅在非 Windows 上断言。
	if runtime.GOOS == "windows" {
		t.Skip("Unix 权限位在 Windows 上无意义，跳过；生产 Linux 上 os.Chmod 0644 生效")
	}
	s := newSaver(t, "")
	rel, err := s.Save([]byte("hello"), "abc", "png", fixed)
	if err != nil {
		t.Fatalf("save: %v", err)
	}
	if rel != "2026/06/abc.png" {
		t.Fatalf("rel = %q", rel)
	}
	abs := filepath.Join(s.RootDir(), "2026", "06", "abc.png")
	info, err := os.Stat(abs)
	if err != nil {
		t.Fatalf("stat written file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o644 {
		t.Fatalf("file mode = %o, want 0644 (nginx 跨用户可读)", got)
	}
}
