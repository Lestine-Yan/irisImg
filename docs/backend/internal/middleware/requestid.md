# `internal/middleware/requestid.go`

请求级追踪中间件。为每个请求生成或透传 `X-Request-Id`，并把它同时写入响应头、`gin.Context` 与 `c.Request.Context()`，使访问日志、业务事件、panic 能经同一个 request id 关联。配套提供 `LogContextFromGin` 把请求上下文里的关联信息集中抽成 `model.LogContext`，供访问日志中间件与业务事件控制器复用。

## 常量

| 常量 | 值 | 说明 |
|------|----|------|
| `ContextKeyRequestID` | `"request_id"` | 存放 request id 的 `gin.Context` 键；业务侧/日志中间件用 `c.GetString(ContextKeyRequestID)` 读取 |
| `HeaderRequestID` | `"X-Request-Id"` | request id 的请求头 / 响应头名称 |

## 函数

### `RequestID() gin.HandlerFunc`

闭包无外部依赖。处理流程：

1. 读 `X-Request-Id` 请求头；非空则**沿用客户端传入的值**（透传），为空则 `uuid.NewString()` 生成。
2. `c.Header(HeaderRequestID, rid)` 回写响应头，客户端可拿到本次请求对应的 id。
3. `c.Set(ContextKeyRequestID, rid)` 写入 `gin.Context`，供 `LogContextFromGin` 及业务代码读取。
4. `c.Request = c.Request.WithContext(logger.ContextWithRequestID(c.Request.Context(), rid))` 写入 `c.Request.Context()`，使 [`logger`](../pkg/logger.md) 的便捷方法（`Info`/`Warn`/`Error` 等）在记录时**自动附加 `request_id` 字段**。
5. `c.Next()`。

两处写入是有意为之：`gin.Context` 面向中间件/控制器（`LogContextFromGin` 取值），`c.Request.Context()` 面向 service/dao 层透传给 [`logger`](../pkg/logger.md)；前者便于组装落库的 `model.Log`，后者让结构化日志无需逐层手传 request id。

### `LogContextFromGin(c *gin.Context) model.LogContext`

从 `gin.Context` 抽取写日志时附带的关联信息，返回 [`model.LogContext`](../model/log.md)。把「上下文取值」集中在一处，避免各调用方重复硬编码键名。

| 字段 | 来源 | 说明 |
|------|------|------|
| `RequestID` | `c.GetString(ContextKeyRequestID)` | 本次请求 id，由 `RequestID` 注入 |
| `Username` | `c.GetString(ContextKeyUsername)` | JWT 鉴权后注入（见 [`auth.md`](auth.md)）；未登录路由为空 |
| `APIKeyID` | `c.Get(ContextKeyAPIKeyID)` | API 密钥鉴权后注入（见 [`apikey.md`](apikey.md)）；类型断言为 `int`，取不到则为 `nil` |
| `ClientIP` | `c.ClientIP()` | gin 解析的客户端 IP |

## 调用关系

```
client ──X-Request-Id(可选)──► RequestID ──┬─► c.Header(X-Request-Id)      回写响应头
                                            ├─► c.Set("request_id")         供 LogContextFromGin 读取
                                            └─► ctx=ContextWithRequestID     供 logger 自动附加字段
                                                       │
                                                       ▼
              Logger / Recovery / 业务控制器 ──► LogContextFromGin(c)
                                                       │
                                                       ▼
                              model.LogContext ──► model.Log / NewEventLog ──► 日志中心
```

`RequestID` 是链路最早执行的中间件之一，确保后续的 [`Logger`](logger.md)（访问日志，`event=http.request`）、[`Recovery`](recovery.md)（panic，`event=panic`）与业务事件控制器（`image.upload` / `apikey.*` 等）拿到的 `RequestID` 一致，从而在日志中心按 `request_id` 过滤即可串起一次请求的全部轨迹。

## 修改建议

- 当前对客户端传入的 `X-Request-Id` **不做校验直接透传**；如需防伪造或超长值，可加长度与字符集校验（如限制为 64 字符内的 uuid/hex），不合法时改用新生成的 uuid。
- request id 仅用 `uuid`，未接入分布式 trace 系统（如 OpenTelemetry）；接入后可在此中间件统一解析/注入 `traceparent`，并把 trace id 一并写入 `LogContext`。
- 新增关联字段（如 user-agent、租户 id）时应在 `LogContextFromGin` 与 [`model.LogContext`](../model/log.md) 同步扩展，避免散落到各调用方重复取值。
