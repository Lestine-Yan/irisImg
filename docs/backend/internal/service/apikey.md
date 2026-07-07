# `internal/service/apikey.go`

API 密钥的业务逻辑层：处理密钥的**签发、管理（列表 / 重命名 / 重置 / 吊销 / 删除）与鉴权**。控制器与中间件把入参原样交过来，本层负责生成明文、调 DAO、判定吊销与权限前置条件，并在删除密钥时级联清理关联图片。

## 类型与变量

### `APIKeyService`

```go
type APIKeyService struct {
    dao      dao.APIKeyDAO
    imageDAO dao.ImageDAO
    saver    *storage.Saver
}
```

- 持有 [`dao.APIKeyDAO`](../dao/dao.md) 接口，由 [`router`](../router/router.md) 注入。
- `imageDAO` 与 `saver` 仅用于 `Delete` 的级联清理（删除该密钥关联的图片文件与记录），其余操作不依赖它们。
- 无可变状态，goroutine 安全。

### sentinel 错误

供上层（中间件 / api）用 `errors.Is` 区分处理：

| 错误 | 含义 |
|------|------|
| `ErrInvalidKeyFormat` | 密钥字符串格式非法（长度 / 字符集不符） |
| `ErrKeyNotFound` | 密钥不存在 |
| `ErrKeyRevoked` | 密钥已被吊销 |
| `ErrInvalidScope` | 请求的权限范围非法 |

## 函数

### `NewAPIKeyService(d dao.APIKeyDAO, imgDAO dao.ImageDAO, saver *storage.Saver) *APIKeyService`

普通构造器。`imgDAO` / `saver` 供 `Delete` 级联清理使用。

### `Create(ctx, *model.CreateAPIKeyRequest) (*model.CreateAPIKeyResponse, error)`

1. 校验 `Scope` 仅为 `readonly` / `readwrite`，否则返回 `ErrInvalidScope`。
2. 调 [`apikey.Generate()`](../pkg/apikey.md) 生成明文、哈希、前缀。
3. `RateLimit < 0` 归一为 0（沿用全局默认）。
4. 调 `dao.Create` 落库（**只存哈希**）。
5. 返回 `CreateAPIKeyResponse`，其中 `Key` 为**一次性明文**。

### `List(ctx) ([]*model.APIKeyInfo, error)`

调 `dao.List`，经私有 `toAPIKeyInfo` 映射为 `APIKeyInfo`（不含明文与哈希）。

### `Revoke(ctx, id int) error`

调 `dao.Revoke`；密钥不存在时把 `dao.ErrNotFound` 映射为 `ErrKeyNotFound`（由控制器映射成 404）。吊销为软删除：密钥仍展示、仍可操作，仅无法通过鉴权。

### `Rename(ctx, id int, name string) (*model.APIKeyInfo, error)`

调 `dao.UpdateName`；`dao.ErrNotFound` → `ErrKeyNotFound`。返回更新后的展示信息。

### `Reset(ctx, id int) (*model.ResetAPIKeyResponse, error)`

1. 调 [`apikey.Generate()`](../pkg/apikey.md) 生成新明文、哈希、前缀。
2. 调 `dao.ResetKey` 写入新哈希/前缀并**清除吊销状态**（重新激活）。
3. 返回 `ResetAPIKeyResponse`，其中 `Key` 为**一次性新明文**；旧明文因哈希被替换而立即失效。
4. `dao.ErrNotFound` → `ErrKeyNotFound`。

### `Delete(ctx, id int) (int, error)`

物理删除密钥并**级联删除其关联图片**（物理文件 + 元信息记录），返回被删除的图片数量。顺序：

1. `dao.GetByID` 确认密钥存在（否则 `ErrKeyNotFound`）。
2. `imageDAO.ListByKeyID` 取该密钥的全部图片路径。
3. 逐个 `saver.Delete(img.StoredPath)` best-effort 删物理文件（失败不阻断）。
4. `imageDAO.DeleteByKeyID` 删图片记录（先删记录以解除外键约束）。
5. `dao.Delete` 删密钥。

> 非事务（删除是低频管理操作）：若删密钥失败，图片已清理但密钥仍在，此时密钥已无关联图片，管理员重试删除不会再受外键约束。

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
middleware.APIKeyAuth ─────►        │   └─► dao.ImageDAO + storage.Saver (Delete 级联清理)
                                    └─► pkg/apikey (Generate / Hash / IsValidFormat)
```

## 修改建议

- 权限矩阵（GET / 非 GET × readonly / readwrite）刻意留在 [`middleware.APIKeyAuth`](../middleware/apikey.md)，service 只负责「这把密钥是否有效」。
- 想支持密钥过期：在 `model.APIKey` / schema 加 `expires_at`，在 `Authenticate` 里加一条判定即可，外部不变。
- 删除级联若要做成事务：给 DAO 透传 ent tx，或把编排上移到一个持 tx 的 service。
