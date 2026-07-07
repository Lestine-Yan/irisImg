# internal/model/apikey.go

API 密钥相关的**跨层数据载体**（实体与请求/响应 DTO）。独立于 Ent 生成的 `ent.ApiKey`：DAO 层负责二者转换（见 [`entdao/apikey.go`](../dao/entdao/apikey.md) 的 `toAPIKeyModel`），使 service / api 层不直接依赖 Ent。

## 常量

```go
const (
    ScopeReadOnly  = "readonly"  // 只读密钥：仅能访问 GET 接口（申请图片）
    ScopeReadWrite = "readwrite" // 读写密钥：可访问 GET 及 POST 接口（添加图片）
)
```

供 service / middleware 做权限判定时引用，避免硬编码字符串。

## 类型

### `APIKey`

密钥实体，字段与 [`ent/schema/apikey.go`](../../ent/schema/apikey.md) 一一对应：`ID`、`Name`、`Prefix`、`Scope`、`KeyHash`、`RateLimit`、`Revoked`、`LastUsedAt`、`CreatedAt`。

- `KeyHash` 带 `json:"-"` 标签：仅在 DAO / service 内部流转，**序列化时始终忽略**，避免泄露哈希。
- 明文密钥不在此结构中（库里只存哈希）。

### `CreateAPIKeyRequest`

创建密钥的请求体。

| 字段 | binding | 说明 |
|------|---------|------|
| `Name` | `required` | 密钥标签 |
| `Scope` | `required,oneof=readonly readwrite` | 权限范围 |
| `RateLimit` | 可选 | 限流阈值（次/分钟），0 表示用全局默认 |

### `CreateAPIKeyResponse`

创建密钥的响应体，比 `APIKey` 多一个 `Key` 字段（明文密钥）。**明文仅在此返回一次**，调用方需自行妥善保存，服务端不再可查。

### `APIKeyInfo`

密钥列表项，不含明文（`Key`）与哈希（`KeyHash`），用于 `GET /apikeys` 列表展示。

### `RenameAPIKeyRequest`

重命名密钥的请求体，仅一个 `Name` 字段（`binding:"required,max=64"`）。

### `ResetAPIKeyResponse`

重置密钥明文后的响应体，与 `CreateAPIKeyResponse` 同构（含一次性 `Key`）。重置会同时取消吊销，故 `Revoked` 恒为 `false`。

### `DestructiveAPIKeyRequest`

吊销 / 删除密钥这类敏感操作的请求体，含 `Username` / `Password`（均 `required`）。后端用 `subtle.ConstantTimeCompare` 校验，作为 JWT 登录态之上的**二次确认**。

### `DeleteAPIKeyResponse`

删除密钥的响应体，`ImagesRemoved` 为被级联删除的图片数量。

## 调用关系

- 被 [`service.APIKeyService`](../service/apikey.md) 构造与消费。
- `KeyHash` 由 [`pkg/apikey`](../pkg/apikey.md) 的 `Hash` 计算后填充。
