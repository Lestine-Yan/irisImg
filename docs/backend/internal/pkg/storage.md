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

- 校验 `root_dir` 非空，`os.MkdirAll` 出存储根目录（提前暴露权限/路径问题）。
- `public_base_url` 允许结尾带或不带 `/`，统一去掉尾斜杠后内部存储，拼接时由 `PublicURL` 补。

### `RelPath(hash, ext string, t time.Time) string`

按时间与文件名计算相对路径 `"YYYY/MM/<hash>.<ext>"`（始终用正斜杠），便于直接用于 URL 拼接。`ext` 不带点、强制小写化；空扩展名兜底为 `bin`。

### `(s *Saver) Save(content []byte, hash, ext string, t time.Time) (string, error)`

- 写到 `<root>/<YYYY>/<MM>/<hash>.<ext>`。
- 目标文件已存在 → 直接复用、不报错（**幂等**，作为 service 秒传判断之外的二次保险）。
- 否则先写到**同目录临时文件**，再 `os.Rename` 原子可见；中途失败会清理临时文件。
- 返回的相对路径**始终使用正斜杠**，方便直接交给 `PublicURL`。

### `(s *Saver) PublicURL(rel string) string`

- `publicBaseURL == ""` → `/imgs/<rel>`（前端 / Nginx 同域反代）。
- 非空 → `<base>/<rel>`（独立图片域名场景）。

### `(s *Saver) RootDir() string`

返回根目录绝对/相对路径，便于测试与诊断。

## 与其它包的关系

```
service.ImageService ──► storage.Saver
                            ├─ Save     (按 hash 写盘)
                            └─ PublicURL(拼对外访问地址)
```

URL 中的 `/imgs` 前缀与生产 Nginx `location /imgs/` → `<root_dir>` 的路径需要保持一致。

## 注意

- 文件存在判定用 `os.Stat`，并发上传同 hash 时也会被 Rename 阶段的兜底再次接住。
- 未尝试做内容校验、缩略图、EXIF 清理；这些放在 service 层或后续扩展。
- 想接入对象存储：实现等价的 `Save` / `PublicURL` 即可，service 持有的是具体类型——必要时再升级为接口。
