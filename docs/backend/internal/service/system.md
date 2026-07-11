# `internal/service/system.go`

系统配置的业务逻辑层：负责把运行时 [`config.Config`](../../config/config.md) 转成对外的**只读系统配置快照**。仅做 `config -> DTO` 的字段映射与脱敏（剥离 dsn 查询参数、排除敏感段），**不读不写任何存储，也不提供任何修改 / 热更新能力**。配置变更需修改 config 文件并重启进程。

## 类型与变量

### `SystemService`

```go
type SystemService struct {
    cfg *config.Config
}
```

- 持有全局 [`config.Config`](../../config/config.md) 的快照引用，由 [`router`](../router/router.md) 通过 `NewSystemService(cfg)` 注入。
- 无 DAO、无存储依赖、无可变状态，goroutine 安全。
- 暴露的唯一方法是只读的 `Config()`，调用方无法触达任何写操作。

## 函数

### `NewSystemService(cfg *config.Config) *SystemService`

构造 `SystemService`，注入全局 config 快照。不做任何初始化副作用（不启动协程、不开通道、不连库）。

### `Config() model.SystemConfigResponse`

返回当前系统配置的只读视图，完成 `config.Config -> model.SystemConfigResponse` 的字段映射。关键处理：

1. **dsn 剥离**：`database.path` 由 `cfg.Database.DSN` 经 `strings.IndexByte(dbPath, '?')` 找到首个 `?`，截取其前的部分得到纯文件路径。例如 `data/irisImg.db?_pragma=busy_timeout(5000)` -> `data/irisImg.db`。若 DSN 不含 `?` 则原样返回，使前端展示纯文件路径而非带 pragma 的连接串。

    ```go
    dbPath := cfg.Database.DSN
    if i := strings.IndexByte(dbPath, '?'); i >= 0 {
        dbPath = dbPath[:i]
    }
    ```

2. **nil 兜底**：`cfg.Storage.AllowedMimeTypes` 为 `nil` 时返回 `[]string{}`，避免 JSON 序列化为 `null`（前端拿到 `null` 而非 `[]` 会增加判空负担）。

    ```go
    mimes := cfg.Storage.AllowedMimeTypes
    if mimes == nil {
        mimes = []string{}
    }
    ```

3. **默认值兜底**：`cfg.APIKey.RateLimitPerMinute <= 0` 时回退 `100`（与 [`ratelimit`](../pkg/ratelimit.md) 包 `NewStore` 的默认一致），`cfg.Storage.MaxUploadSizeMB <= 0` 时回退 `20`（与 [`image.go`](image.md) 的 `ImageService` 默认一致）。0 / 负数表示未配置，若原样透传会让前端展示误导性的 `0`；回退后展示的是实际生效阈值而非 0。

    ```go
    rateLimit := cfg.APIKey.RateLimitPerMinute
    if rateLimit <= 0 {
        rateLimit = 100 // 与 ratelimit.NewStore 的默认一致
    }
    maxUpload := cfg.Storage.MaxUploadSizeMB
    if maxUpload <= 0 {
        maxUpload = 20 // 与 ImageService 的默认一致
    }
    ```

4. **字段直传**：`Server`（`Host` / `Port`）、`APIKey`（兜底后的 `RateLimitPerMinute` / `HTTPSOnly`）、`Storage`（`RootDir` / `PublicBaseURL` / 兜底后的 `MaxUploadSizeMB` / 处理后的 `AllowedMimeTypes`）与 `Database.Driver` 从 config 透传；其中 `RateLimitPerMinute` 与 `MaxUploadSizeMB` 透传的是上一步兜底后的值，而非 config 原值。
5. **脱敏**：`config.Auth` 段（含 `password` 与 `jwt.secret`）**不参与映射**，响应体中不出现这些机密。

由于是纯内存映射、无 IO，本方法不返回 error。

## 与其它包的关系

```
api.SystemAPI ──► service.SystemService ──► config.Config (只读快照)
router ──(NewSystemService)──────────►│
                                      └─► model.SystemConfigResponse (DTO)
```

- 不依赖 `dao` / `storage` / `logger`，是全后端最轻量的 service。
- 控制器层细节见 [`../api/system.md`](../api/system.md)；DTO 字段定义见 [`../model/system.md`](../model/system.md)。

## 测试

`system_test.go` 以表驱动测试覆盖 `Config()` 的关键字段映射与兜底逻辑：

- `TestSystemService_Config`：端到端验证 `config -> DTO` 映射--host/port 透传、dsn 剥离 `?` 得纯路径、`0` 值回退生效默认（限速 `100` / 上传 `20` MiB）、`nil` `AllowedMimeTypes` 兜底为非 nil 空切片（避免 JSON `null`）。
- `TestSystemService_ConfigDSNStripping`：表驱动覆盖 dsn 剥离 `?` 的四条边界--无 `?` / `?` 在中部 / `?` 在首位 / 空 dsn。
- `TestSystemService_ConfigExplicitValues`：显式非零值（限速 `30` / 上传 `5` MiB / 非空 MIME 列表 / `HTTPSOnly=true`）不被默认覆盖。

## 修改建议

- 该服务刻意保持**只读**：如需在线修改配置，应新增独立的写服务并配套审计日志，不要在 `SystemService` 上直接加 setter。
- 当前 DSN 剥离只处理首个 `?`；若未来 DSN 出现更复杂结构（如带 scheme 的连接串），需同步调整剥离逻辑并在本文档标注。
- 若新增可暴露的配置段，务必先评估是否含敏感字段；`Auth` 段的脱敏边界应长期保持，不要为「前端展示完整度」而放宽。
