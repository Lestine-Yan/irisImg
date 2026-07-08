# `internal/middleware/logger.go`

zap 结构化访问日志中间件。每个请求（健康检查除外）在 `c.Next()` 后用 zap 输出结构化访问日志，并异步落库一条 `event=http.request` 日志，供日志中心查询。

## 函数

### `Logger(l *logger.Logger, svc *service.LogService) gin.HandlerFunc`

闭包持有 [`*logger.Logger`](../pkg/logger.md) 与 [`*service.LogService`](../service/log.md)，由 [`router`](../router/router.md) 注入。流程：

- 跳过 `/api/v1/ping` 健康检查，直接 `c.Next()` 返回，减噪。
- 进入前记 `start = time.Now()`、`path`；`c.Next()` 让后续处理器跑完。
- 取 `status = c.Writer.Status()`、`duration = time.Since(start)`，按状态码推导级别 `levelFromStatus`。
- 用 [`LogContextFromGin`](requestid.md)`(c)` 抽取 `request_id / username / api_key_id / client_ip`（定义在 `requestid.go`）。

zap 输出消息 `http.request`，按 level 分发到 `l.Info` / `l.Warn` / `l.Error`，类型化字段：

| zap 字段 | 来源 | 备注 |
|----------|------|------|
| `method` | `c.Request.Method` | |
| `path` | `c.Request.URL.Path` | |
| `status` | `c.Writer.Status()` | int |
| `duration` | `time.Since(start)` | `zap.Duration` |
| `bytes` | `c.Writer.Size()` | 响应字节数 |
| `username` | `lc.Username` | 非空才追加 |
| `api_key_id` | `lc.APIKeyID` | 非 nil 才追加 |
| `request_id` | — | 不在此显式写；`RequestID` 中间件已把 id 写进 `c.Request.Context()`，`logger.Logger` 的 `Info/Warn/Error` 会自动提取并附加 |

落库（`svc != nil` 时）异步写一条 [`model.Log`](../model/log.md)，`Event = model.EventHTTPRequest`，含 `Timestamp / Level / Method / Path / Status / DurationMs / ClientIP / RequestID / APIKeyID / Username`。

### `levelFromStatus(status int) string`

按 HTTP 状态码推导日志级别：

| status | 返回 |
|--------|------|
| ≥ 500 | `model.LevelError` |
| ≥ 400 | `model.LevelWarn` |
| 其它 | `model.LevelInfo` |

## 与其它包的关系

```
client ──X-Request-Id──► RequestID ──► Logger ──► c.Next() ──► handlers
                                     │
                                     ├─► LogContextFromGin(c) ──► {request_id, username, api_key_id, client_ip}
                                     ├─► logger.Logger.{Info,Warn,Error}  (zap 类型化字段；request_id 自动附加)
                                     └─► service.LogService.Record ──► dao ──► logs 表
```

## 修改建议

- 已落地：zap 结构化日志、trace id（`X-Request-Id` 透传 + ctx 自动附加）、客户端 IP、访问日志落库（`LogService.Record`）。
- 后续可加：慢请求阈值告警（`duration` 超阈值升级为 warn）、按路由采样、请求体大小（需在 `c.Next` 前读 body，谨慎）。
- 不要在这里输出敏感数据（如登录请求体），日志会泄漏密码。
