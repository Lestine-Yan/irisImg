# internal/model/system.go

系统配置只读接口（`GET /system/config`）的**响应 DTO**。把运行时 [`config.Config`](../../config/config.md) 的非敏感字段映射为对外可暴露的只读快照，供前端「系统配置」页面展示。该接口为只读快照，配置变更需修改 config 文件并重启，前端不做在线编辑。

## 类型

### `SystemConfigResponse`

接口的顶层响应体，由四个子视图聚合而成：

| 字段 | 类型 | json tag | 说明 |
|------|------|----------|------|
| `Server` | `ServerConfigView` | `server` | 服务监听信息 |
| `Database` | `DatabaseConfigView` | `database` | 数据库位置信息 |
| `APIKey` | `APIKeyConfigView` | `apikey` | API 密钥全局开关与默认阈值 |
| `Storage` | `StorageConfigView` | `storage` | 图片存储相关参数 |

> **脱敏设计**：[`config.Config`](../../config/config.md) 中的 `Auth` 段（`username` / `password` / `jwt.secret`）**刻意不在此结构中暴露**，避免把管理员口令与签名密钥下发到前端。`Database` 仅暴露驱动与纯文件路径，`DSN` 中的连接参数（如 `?_pragma=busy_timeout(5000)`）由 service 层剥离（见 [`service.SystemService`](../service/system.md)）。`App` / `Logger` 段当前亦不暴露。

### `ServerConfigView`

暴露服务监听信息。

| 字段 | json tag | 说明 |
|------|----------|------|
| `Host` | `host` | 监听地址，如 `"0.0.0.0"` |
| `Port` | `port` | 监听端口，如 `8080` |

### `DatabaseConfigView`

暴露数据库位置信息。

| 字段 | json tag | 说明 |
|------|----------|------|
| `Driver` | `driver` | 驱动，当前固定 `sqlite` |
| `Path` | `path` | 数据库文件路径（由 service 层从 `database.dsn` 剥离首个 `?` 之后的查询参数得到，非 DSN 原文） |

### `APIKeyConfigView`

暴露 API 密钥相关的全局开关与默认阈值。

| 字段 | json tag | 说明 |
|------|----------|------|
| `RateLimitPerMinute` | `rate_limit_per_minute` | 全局默认限流阈值（次/分钟），单密钥 `rate_limit` 为 0 时回退到此值 |
| `HTTPSOnly` | `https_only` | 是否强制密钥敏感接口走 HTTPS，`false` 时前端应给出安全警告 |

### `StorageConfigView`

暴露图片存储相关参数。

| 字段 | json tag | 说明 |
|------|----------|------|
| `RootDir` | `root_dir` | 图片落盘根目录 |
| `PublicBaseURL` | `public_base_url` | 图片对外访问基址，空表示走相对路径 `/imgs/` |
| `MaxUploadSizeMB` | `max_upload_size_mb` | 单次上传上限（MiB） |
| `AllowedMimeTypes` | `allowed_mime_types` | 允许上传的 MIME 白名单；service 层在配置为 `nil` 时兜底为空切片，避免 JSON 序列化为 `null` |

## 调用关系

- 由 [`service.SystemService`](../service/system.md) 在 `Config()` 中构造并填充。
- 经 [`api.SystemAPI`](../api/system.md) 的 `Config` handler 经 [`response.Success`](../pkg/response.md) 下发给前端。
- 字段来源为 [`config.Config`](../../config/config.md)；`Auth` 段不参与映射。
