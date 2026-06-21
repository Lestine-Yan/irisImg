# `internal/middleware/apikey.go`

API 密钥鉴权中间件。挂在图片接口分组上，校验请求头里的密钥并把密钥 ID 注入 `gin.Context`。独立于 JWT 鉴权（见 [`auth.md`](auth.md)）。

## 常量

| 常量 | 值 | 说明 |
|------|----|------|
| `HeaderAPIKey` | `"X-API-Key"` | 携带 API 密钥的请求头名称 |
| `ContextKeyAPIKeyID` | `"api_key_id"` | 上下文键；业务侧 `c.GetInt(ContextKeyAPIKeyID)` 取出当前密钥 ID，用于落库「图片由哪个密钥添加」 |

## 函数

### `APIKeyAuth(svc *service.APIKeyService, limiter *ratelimit.Store) gin.HandlerFunc`

闭包持有 [`service.APIKeyService`](../service/apikey.md) 与 [`ratelimit.Store`](../pkg/ratelimit.md)，由 [`router`](../router/router.md) 注入。按顺序校验，任一步失败即 `c.Abort()`：

| 步骤 | 触发条件 | HTTP | code | message 示例 |
|------|---------|------|------|--------------|
| 1 | `X-API-Key` 缺失 / 为空 | 401 | `CodeAPIKeyMissing`(40110) | `缺少 X-API-Key 请求头` |
| 2 | 格式非法（`ErrInvalidKeyFormat`） | 401 | `CodeAPIKeyInvalid`(40120) | `密钥格式非法` |
| 3a | 不存在（`ErrKeyNotFound`） | 401 | `CodeAPIKeyInvalid` | `密钥不存在` |
| 3b | 已吊销（`ErrKeyRevoked`） | 401 | `CodeAPIKeyInvalid` | `密钥已吊销` |
| 3c | 其它内部错误 | 500 | `CodeServerError` | `密钥校验失败` |
| 4 | 非 GET 请求且非 `readwrite` 密钥 | 403 | `CodeForbidden`(40300) | `只读密钥无权访问该接口` |
| 5 | 触发限流 | 429 | `CodeTooManyRequests`(42900) | `请求过于频繁，请稍后再试` |

步骤 2 & 3 调 `svc.Authenticate`（格式 / 查库 / 吊销）；步骤 4 是权限矩阵：**GET 请求任意有效密钥可访问，非 GET（POST 等）必须为 readwrite**；步骤 5 调 `limiter.Allow(key.ID, key.RateLimit)`。

通过后：

```go
c.Set(ContextKeyAPIKeyID, key.ID)
_ = svc.Touch(c.Request.Context(), key.ID) // 尽力更新最近使用时间，失败不阻断
c.Next()
```

## 与其它包的关系

```
client ──X-API-Key──► APIKeyAuth ──► service.Authenticate ──► dao.GetByHash
                          │
                          ├─► ratelimit.Store.Allow
                          └─► c.Set("api_key_id", id) ──► api.ImageAPI
```

## 注意

- 错误码刻意区分：缺失（40110）与无效（40120）分开，便于前端/调用方分别提示「没带 key」与「key 不对」。
- 权限与限流判定在本中间件而非 service：因为需要 HTTP 方法与令牌桶上下文。
- `Touch` 失败被忽略：最近使用时间是观测信息，不应因写库失败而拒绝合法请求。
