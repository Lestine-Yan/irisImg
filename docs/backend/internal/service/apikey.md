# `internal/service/apikey.go`

API 密钥的业务逻辑层：处理密钥的**签发、管理（列表 / 吊销 / 更新使用时间）与鉴权**。控制器与中间件把入参原样交过来，本层负责生成明文、调 DAO、判定吊销与权限前置条件。

## 类型与变量

### `APIKeyService`

```go
type APIKeyService struct {
    dao dao.APIKeyDAO
}
```

持有 [`dao.APIKeyDAO`](../dao/dao.md) 接口，由 [`router`](../router/router.md) 注入；无可变状态，goroutine 安全。

### sentinel 错误

供上层（中间件 / api）用 `errors.Is` 区分处理：

| 错误 | 含义 |
|------|------|
| `ErrInvalidKeyFormat` | 密钥字符串格式非法（长度 / 字符集不符） |
| `ErrKeyNotFound` | 密钥不存在 |
| `ErrKeyRevoked` | 密钥已被吊销 |
| `ErrInvalidScope` | 请求的权限范围非法 |

## 函数

### `NewAPIKeyService(d dao.APIKeyDAO) *APIKeyService`

普通构造器。

### `Create(ctx, *model.CreateAPIKeyRequest) (*model.CreateAPIKeyResponse, error)`

1. 校验 `Scope` 仅为 `readonly` / `readwrite`，否则返回 `ErrInvalidScope`。
2. 调 [`apikey.Generate()`](../pkg/apikey.md) 生成明文、哈希、前缀。
3. `RateLimit < 0` 归一为 0（沿用全局默认）。
4. 调 `dao.Create` 落库（**只存哈希**）。
5. 返回 `CreateAPIKeyResponse`，其中 `Key` 为**一次性明文**。

### `List(ctx) ([]*model.APIKeyInfo, error)`

调 `dao.List`，映射为 `APIKeyInfo`（不含明文与哈希）。

### `Revoke(ctx, id int) error`

调 `dao.Revoke`；密钥不存在时返回 `dao.ErrNotFound`（由控制器映射成 404）。

### `Touch(ctx, id int) error`

把最近使用时间更新为 `time.Now()`，供鉴权通过后尽力调用（失败不阻断主流程）。

### `Authenticate(ctx, plaintext string) (*model.APIKey, error)`

鉴权核心：**格式校验 → 查库 → 吊销判定**。

```go
if !apikeypkg.IsValidFormat(plaintext) { return nil, ErrInvalidKeyFormat }
key, err := s.dao.GetByHash(ctx, apikeypkg.Hash(plaintext))
// dao.ErrNotFound -> ErrKeyNotFound
if key.Revoked { return nil, ErrKeyRevoked }
return key, nil
```

注意：**不做权限（scope）与限流判定**，那两步在中间件里完成（中间件需知道 HTTP 方法与令牌桶）。

## 与其它包的关系

```
api.APIKeyAPI ────────────► service.APIKeyService ──► dao.APIKeyDAO
middleware.APIKeyAuth ─────►        │
                                    └─► pkg/apikey (Generate / Hash / IsValidFormat)
```

## 修改建议

- 权限矩阵（GET / 非 GET × readonly / readwrite）刻意留在 [`middleware.APIKeyAuth`](../middleware/apikey.md)，service 只负责「这把密钥是否有效」。
- 想支持密钥过期：在 `model.APIKey` / schema 加 `expires_at`，在 `Authenticate` 里加一条判定即可，外部不变。
