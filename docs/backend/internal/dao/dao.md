# internal/dao/dao.go

持久化访问的**抽象接口**定义包。业务层（service）只依赖这里的接口，不感知具体存储后端，因此可在不改业务逻辑的前提下替换底层实现。当前唯一实现位于子包 [`entdao`](./entdao/db.md)（Ent + modernc.org/sqlite）。

## 接口：ImageDAO

抽象图片元信息的持久化操作，数据载体为 [`model.Image`](../model/image.md)。

| 方法 | 说明 |
|------|------|
| `Create(ctx, *model.Image) (*model.Image, error)` | 落库一条记录，回填自增 ID 与创建时间 |
| `GetByID(ctx, id int) (*model.Image, error)` | 按主键查询，未找到返回 `ErrNotFound` |
| `GetByHash(ctx, hash string) (*model.Image, error)` | 按内容哈希查询（秒传 / 去重），未找到返回 `ErrNotFound` |
| `List(ctx, offset, limit int) ([]*model.Image, int, error)` | 按创建时间倒序分页，返回条目与总数 |
| `Delete(ctx, id int) error` | 按主键删除，未找到返回 `ErrNotFound` |

## 错误

统一错误见 [`errors.go`](./errors.md)：各实现需将底层「记录不存在」转换为 `dao.ErrNotFound`。

## 调用关系

- 实现：[`internal/dao/entdao`](./entdao/db.md)
- 注入：`cmd/server/main.go` 构造实现后通过 `router.New` 注入（见 [`cmd/server.md`](../../cmd/server.md)、[`router.md`](../router/router.md)）。
