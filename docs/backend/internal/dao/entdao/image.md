# internal/dao/entdao/image.go

[`dao.ImageDAO`](../dao.md) 的 Ent 实现。

## 类型

- `imageDAO`：持有 `*ent.Client` 的非导出实现类型。
- `NewImageDAO(client *ent.Client) dao.ImageDAO`：构造函数。
- 编译期断言 `var _ dao.ImageDAO = (*imageDAO)(nil)` 保证接口一致。

## 方法

逐一实现 `ImageDAO` 接口：`Create` / `GetByID` / `GetByHash` / `List` / `Delete`。

- `List`：先 `Count` 取总数，再按 `created_at` 倒序（`ent.Desc`）分页；`offset` / `limit` 为正才生效。
- 查询单条用 ent 生成的 `image.HashEQ` 等谓词。

## 错误与转换

- `wrapErr`：`ent.IsNotFound(err)` → [`dao.ErrNotFound`](../errors.md)，屏蔽 Ent 错误类型。
- `toModel(*ent.Image) *model.Image`：把 Ent 实体转换为跨层的 [`model.Image`](../../model/image.md)，使 service / api 不直接依赖 Ent。

## 测试

`image_test.go` 在 `t.TempDir()` 打开真实 SQLite（纯 Go 驱动、离线、无 CGO），覆盖创建/查询、未找到、列表分页与删除。
