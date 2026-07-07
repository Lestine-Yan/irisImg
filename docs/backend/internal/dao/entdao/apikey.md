# internal/dao/entdao/apikey.go

[`dao.APIKeyDAO`](../dao.md) 的 Ent 实现。

## 类型

- `apiKeyDAO`：持有 `*ent.Client` 的非导出实现类型。
- `NewAPIKeyDAO(client *ent.Client) dao.APIKeyDAO`：构造函数。
- 编译期断言 `var _ dao.APIKeyDAO = (*apiKeyDAO)(nil)` 保证接口一致。

## 方法

逐一实现 `APIKeyDAO` 接口：

| 方法 | 实现要点 |
|------|---------|
| `Create` | `client.ApiKey.Create()` 写入 name / key_hash / prefix / scope / rate_limit；`scope` 用生成的 `apikey.Scope(key.Scope)` 转枚举 |
| `GetByHash` | `apikey.KeyHashEQ(hash)` + `First`，用于鉴权按哈希查找 |
| `GetByID` | `client.ApiKey.Get(ctx, id)` |
| `List` | 按 `created_at` 倒序（`ent.Desc`）返回全部密钥 |
| `Revoke` | `UpdateOneID(id).SetRevoked(true)` |
| `UpdateName` | `UpdateOneID(id).SetName(name).Save(ctx)`，返回更新后的实体 |
| `ResetKey` | `UpdateOneID(id).SetKeyHash(...).SetPrefix(...).SetRevoked(false).Save(ctx)`，重置明文并取消吊销 |
| `Delete` | `DeleteOneID(id).Exec(ctx)`，物理删除 |
| `TouchLastUsed` | `UpdateOneID(id).SetLastUsedAt(t)`，鉴权通过后尽力更新最近使用时间 |

## 错误与转换

- 复用同包 [`image.go`](image.md) 的 `wrapErr`：`ent.IsNotFound(err)` → [`dao.ErrNotFound`](../errors.md)，屏蔽 Ent 错误类型。`Create` / `List` 直接透传底层错误（无「不存在」语义）。
- `toAPIKeyModel(*ent.ApiKey) *model.APIKey`：把 Ent 实体转换为跨层的 [`model.APIKey`](../../model/apikey.md)，回填 `KeyHash`（仅内部流转，JSON 忽略）。

## 调用关系

被 [`service.APIKeyService`](../../service/apikey.md) 依赖（通过 `dao.APIKeyDAO` 接口）；由 [`cmd/server/main.go`](../../../cmd/server.md) 构造后经 [`router.New`](../../router/router.md) 注入。
