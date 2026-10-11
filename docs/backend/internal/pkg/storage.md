# `internal/pkg/storage/storage.go`

> 文档路径约定：`internal/pkg/storage/storage.go` → `docs/backend/internal/pkg/storage.md`（去掉内层同名目录）。

提供图片/音频二进制的**本地文件落盘**与**对外访问 URL 拼接**能力。设计成「薄、可替换」的小工具，service 层只依赖导出的 `Saver`，后续要换成对象存储（S3 / OSS / COS）时只需提供等价实现，不必动业务逻辑。同一个 `Saver` 类型可服务不同的静态前缀：图片用 `NewSaver`（`/imgs`），音频用 `NewAudioSaver`（`/audios`），二者仅 rootDir / publicBaseURL / URL 前缀不同，落盘与原子写逻辑完全共享。

## 目录与文件名约定

落盘路径：

```
<root_dir>/<YYYY>/<MM>/<sha256>.<ext>
```

- `root_dir` 来自配置 `storage.root_dir`（图片）/ `storage.audio_root_dir`（音频），相对路径相对于进程 cwd，部署时建议改成绝对路径。
- 年月分桶避免单目录文件过多。
- 文件名采用内容的 SHA256 哈希，**天然唯一、天然支持秒传**（同一文件必然落到同一路径）。
- 扩展名由 service 层根据嗅探出的真实 MIME / 格式推导，全部小写，不信任原始文件名后缀。
- **例外**：音频内嵌封面用 `SaveFlat` 平铺到 `<root>/covers/<sha256>.<ext>`（内容寻址，同一专辑封面跨音频共享天然去重）。

## 类型

### `Saver`

```go
type Saver struct {
    rootDir       string
    publicBaseURL string
    relPrefix     string   // 相对 URL 前缀："/imgs" 或 "/audios"，构造时决定
}
```

字段在构造后只读，goroutine 安全。

## 函数

### `NewSaver(cfg config.StorageConfig) (*Saver, error)`

图片构造器（`storage.root_dir` / `storage.public_base_url`，前缀 `/imgs`）。

### `NewAudioSaver(cfg config.StorageConfig) (*Saver, error)`

音频构造器（`storage.audio_root_dir` / `storage.audio_public_base_url`，前缀 `/audios`），语义与 `NewSaver` 完全一致。

### `newSaverWith(cfgKey, root, publicBaseURL, prefix string) (*Saver, error)`

两构造器的共用内部实现：

- 校验 `root` 非空。
- **启动期 fail-fast 拒绝危险 root**：调用 `guardRootDir` 拒绝把存储根目录配成后端工作目录本身或其祖先（含 `.` / `..` / `/` / 工作目录父级）。这类配置会让静态服务把 `config.yaml` / `irisImg.db` / 源码暴露给未认证访客；即便 [`serveFiles`](../router/static.md) 有扩展名白名单兜底，这里把误配扼杀在启动阶段、错误更明确（报错带配置键名与前缀，音频/图片各自可辨）。
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

- 写到 `<root>/<YYYY>/<MM>/<hash>.<ext>`（原子写核心抽取为 `saveRel`，与 `SaveFlat` 共享）。
- 目标文件已存在 → 直接复用、不报错（**幂等**，作为 service 秒传判断之外的二次保险）。
- 否则先写到**同目录临时文件**，再 `os.Rename` 原子可见；中途失败会清理临时文件。
- **落盘权限 0644**：`os.CreateTemp` 以 0600 创建临时文件、`os.Rename` 只改名不改权限，故 Rename 后显式 `os.Chmod(abs, 0o644)`，保证生产环境 Nginx（`www` 用户，作为 other）能跨用户读取，否则会 403。秒传兜底分支（Rename 失败但文件已存在）也幂等 Chmod 一次，纠正旧版本落盘的 0600 历史文件。
- 返回的相对路径**始终使用正斜杠**，方便直接交给 `PublicURL`。

### `(s *Saver) SaveFlat(rel string, content []byte) error`

把内容写到调用方给定的相对路径（相对 rootDir、正斜杠），**不做年月分桶**。供音频内嵌封面这类「内容寻址、跨记录共享」的文件使用（`covers/<sha256>.<ext>`）。幂等语义与 `Save` 一致（已存在直接复用）；`rel` 会做 `path.Clean` 并拒绝绝对路径与 `..` 逃逸（调用方是内部代码，纯防御性兜底）。

### `(s *Saver) PublicURL(rel string) string`

- `publicBaseURL == ""` → `<relPrefix>/<rel>`（图片 `/imgs/...`，音频 `/audios/...`；前端 / Nginx 同域反代）。
- 非空 → `<base>/<rel>`（独立域名场景）。`base` 已由构造器规范化为带协议（裸域名补 `https://`），故此处产出的总是浏览器可直接加载的绝对地址或同域相对地址。

### `(s *Saver) Delete(rel string) error`

- 按 `Save` 返回的相对路径（正斜杠）删除物理文件，内部转回平台分隔符后拼到 `rootDir` 之下。
- 文件不存在视为已删除（**幂等**），不报错——供删除密钥/批量删除时 best-effort 清理关联文件。
- 空路径直接返回 nil。
- **注意**：音频内嵌封面（covers/ 下内容寻址文件）**刻意不走删除**——同一封面被多个音频共享，删除任一音频不应影响其余音频的封面展示；孤儿清扫是后续可选的运维操作。

### `(s *Saver) RootDir() string`

返回根目录绝对/相对路径，便于测试与诊断。

## 与其它包的关系

```
service.ImageService ──► storage.Saver（NewSaver，/imgs）
service.AudioService ──► storage.Saver（NewAudioSaver，/audios）
                            ├─ Save     (按 hash 写盘，年月分桶)
                            ├─ SaveFlat (平铺写盘，音频封面 covers/)
                            ├─ Delete   (删密钥/批量删除时清理关联内容文件)
                            └─ PublicURL(拼对外访问地址，前缀由构造器决定)
service.APIKeyService ──► 两个 Saver（Delete，级联清理图片与音频文件）
```

URL 中的 `/imgs`、`/audios` 前缀与生产 Nginx `location /imgs/`、`location /audios/` → 对应根目录的路径需要保持一致。

## 注意

- 文件存在判定用 `os.Stat`，并发上传同 hash 时也会被 Rename 阶段的兜底再次接住。
- 未尝试做内容校验、缩略图、EXIF 清理；这些放在 service 层或后续扩展。
- 想接入对象存储：实现等价的 `Save` / `SaveFlat` / `PublicURL` 即可，service 持有的是具体类型——必要时再升级为接口。
