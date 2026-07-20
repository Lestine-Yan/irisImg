# `internal/pkg/storage/storage.go`

> 文档路径约定：`internal/pkg/storage/storage.go` → `docs/backend/internal/pkg/storage.md`（去掉内层同名目录）。

提供图片二进制的**本地文件落盘**与**对外访问 URL 拼接**能力。设计成「薄、可替换」的小工具，service 层只依赖导出的 `Saver`，后续要换成对象存储（S3 / OSS / COS）时只需提供等价实现，不必动业务逻辑。

## 目录与文件名约定

落盘路径：

```
<root_dir>/<YYYY>/<MM>/<sha256>.<ext>
```

- `root_dir` 来自配置 `storage.root_dir`（相对路径相对于进程 cwd，部署时建议改成绝对路径）。
- 年月分桶避免单目录文件过多。
- 文件名采用图片的 SHA256 哈希，**天然唯一、天然支持秒传**（同一张图必然落到同一路径）。
- 扩展名由 service 层根据嗅探出的真实 MIME 推导，全部小写，不信任原始文件名后缀。

## 类型

### `Saver`

```go
type Saver struct {
    rootDir       string
    publicBaseURL string
}
```

字段在构造后只读，goroutine 安全。

## 函数

### `NewSaver(cfg config.StorageConfig) (*Saver, error)`

- 校验 `root_dir` 非空。
- **启动期 fail-fast 拒绝危险 root**：调用 `guardRootDir` 拒绝把 `root_dir` 配成后端工作目录本身或其祖先（含 `.` / `..` / `/` / 工作目录父级）。这类配置会让 `/imgs` 静态服务把 `config.yaml` / `irisImg.db` / 源码暴露给未认证访客；即便 [`serveImages`](../router/static.md) 有扩展名白名单兜底，这里把误配扼杀在启动阶段、错误更明确。
- `os.MkdirAll` 出存储根目录（提前暴露权限/路径问题）。
- `public_base_url` 允许结尾带或不带 `/`，统一去掉尾斜杠后内部存储，拼接时由 `PublicURL` 补。
- **裸域名容错**：`public_base_url` 非空且不含 `://`（如 `img.example.com/imgs`）时自动补 `https://` 前缀，避免前端把无协议值当相对路径解析（拼成 `/img.example.com/imgs/...`）。已带 `http://`/`https://` 的原样保留。

### `guardRootDir(root string) error`

判定 `root` 是否为后端工作目录（cwd）本身或其祖先：

- `filepath.Rel(absRoot, absCwd)` 得到「cwd 相对 root 的路径」；不以 `..` 开头（含 `.`）说明 cwd 落在 root 之内或等于 root -> root 是 cwd 本身/祖先 -> 返回 error。
- 取不到 cwd 或无法计算相对路径（如跨盘符）时保守放行，交由扩展名白名单兜底。
- 安全形态均放行：cwd 之下的独立子目录（默认 `data/imgs`）、cwd 之外的专用目录（生产 `/var/lib/irisImg/imgs`）。

### `RelPath(hash, ext string, t time.Time) string`

按时间与文件名计算相对路径 `"YYYY/MM/<hash>.<ext>"`（始终用正斜杠），便于直接用于 URL 拼接。`ext` 不带点、强制小写化；空扩展名兜底为 `bin`。

### `(s *Saver) Save(content []byte, hash, ext string, t time.Time) (string, error)`

- 写到 `<root>/<YYYY>/<MM>/<hash>.<ext>`。
- 目标文件已存在 → 直接复用、不报错（**幂等**，作为 service 秒传判断之外的二次保险）。
- 否则先写到**同目录临时文件**，再 `os.Rename` 原子可见；中途失败会清理临时文件。
- **落盘权限 0644**：`os.CreateTemp` 以 0600 创建临时文件、`os.Rename` 只改名不改权限，故 Rename 后显式 `os.Chmod(abs, 0o644)`，保证生产环境 Nginx（`www` 用户，作为 other）能跨用户读取，否则会 403。秒传兜底分支（Rename 失败但文件已存在）也幂等 Chmod 一次，纠正旧版本落盘的 0600 历史文件。
- 返回的相对路径**始终使用正斜杠**，方便直接交给 `PublicURL`。

### `(s *Saver) PublicURL(rel string) string`

- `publicBaseURL == ""` → `/imgs/<rel>`（前端 / Nginx 同域反代）。
- 非空 → `<base>/<rel>`（独立图片域名场景）。`base` 已由 `NewSaver` 规范化为带协议（裸域名补 `https://`），故此处产出的总是浏览器可直接加载的绝对地址或同域相对地址。

### `(s *Saver) Delete(rel string) error`

- 按 `Save` 返回的相对路径（正斜杠）删除物理文件，内部转回平台分隔符后拼到 `rootDir` 之下。
- 文件不存在视为已删除（**幂等**），不报错——供删除密钥时 best-effort 清理关联图片文件。
- 空路径直接返回 nil。

### `(s *Saver) RootDir() string`

返回根目录绝对/相对路径，便于测试与诊断。

## 与其它包的关系

```
service.ImageService ──► storage.Saver
                            ├─ Save     (按 hash 写盘)
                            ├─ Delete   (删密钥时清理关联图片文件)
                            └─ PublicURL(拼对外访问地址)
service.APIKeyService ──► storage.Saver（Delete，级联清理）
```

URL 中的 `/imgs` 前缀与生产 Nginx `location /imgs/` → `<root_dir>` 的路径需要保持一致。

## 注意

- 文件存在判定用 `os.Stat`，并发上传同 hash 时也会被 Rename 阶段的兜底再次接住。
- 未尝试做内容校验、缩略图、EXIF 清理；这些放在 service 层或后续扩展。
- 想接入对象存储：实现等价的 `Save` / `PublicURL` 即可，service 持有的是具体类型——必要时再升级为接口。
