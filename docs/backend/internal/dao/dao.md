# internal/dao/dao.go

持久化访问的**抽象接口**定义包。业务层（service）只依赖这里的接口，不感知具体存储后端，因此可在不改业务逻辑的前提下替换底层实现。当前唯一实现位于子包 [`entdao`](./entdao/db.md)（Ent + modernc.org/sqlite）。

## 接口：ImageDAO

抽象图片元信息的持久化操作，数据载体为 [`model.Image`](../model/image.md)。

| 方法 | 说明 |
|------|------|
| `Create(ctx, *model.Image) (*model.Image, error)` | 落库一条记录，回填自增 ID 与创建时间 |
| `GetByID(ctx, id int) (*model.Image, error)` | 按主键查询，未找到返回 `ErrNotFound` |
| `GetByHash(ctx, hash string) (*model.Image, error)` | 按内容哈希查询（秒传 / 去重），未找到返回 `ErrNotFound` |
| `List(ctx, q model.ImageListQuery) ([]*model.Image, int, error)` | 按 `ImageListQuery` 过滤（可选 key_id）/ 排序（asc/desc）/ 分页，返回条目与符合过滤条件的总数 |
| `ListByKeyID(ctx, keyID int) ([]*model.Image, error)` | 返回指定密钥关联的全部图片（不分页），供删除密钥时级联清理使用 |
| `Delete(ctx, id int) error` | 按主键删除，未找到返回 `ErrNotFound` |
| `DeleteByKeyID(ctx, keyID int) (int, error)` | 批量删除指定密钥关联的全部图片记录，返回删除条数 |

## 错误

统一错误见 [`errors.go`](./errors.md)：各实现需将底层「记录不存在」转换为 `dao.ErrNotFound`。

## 接口：APIKeyDAO

抽象 API 密钥的持久化操作，数据载体为 [`model.APIKey`](../model/apikey.md)（库里只存哈希）。

| 方法 | 说明 |
|------|------|
| `Create(ctx, *model.APIKey) (*model.APIKey, error)` | 落库一把密钥（存哈希），回填自增 ID 与创建时间 |
| `GetByHash(ctx, hash string) (*model.APIKey, error)` | 按密钥哈希查询（鉴权用），未找到返回 `ErrNotFound` |
| `GetByID(ctx, id int) (*model.APIKey, error)` | 按主键查询，未找到返回 `ErrNotFound` |
| `List(ctx) ([]*model.APIKey, error)` | 按创建时间倒序返回全部密钥 |
| `Revoke(ctx, id int) error` | 标记为已吊销，未找到返回 `ErrNotFound` |
| `UpdateName(ctx, id int, name string) (*model.APIKey, error)` | 修改密钥标签，未找到返回 `ErrNotFound` |
| `ResetKey(ctx, id int, keyHash, prefix string) (*model.APIKey, error)` | 重置明文：写入新哈希/前缀并清除吊销，未找到返回 `ErrNotFound` |
| `Delete(ctx, id int) error` | 物理删除，未找到返回 `ErrNotFound`；调用方需先清理关联图片以免外键约束失败 |
| `TouchLastUsed(ctx, id int, t time.Time) error` | 更新最近使用时间 |

实现见 [`entdao/apikey.go`](./entdao/apikey.md)。特性级说明见 [`APIKEY.md`](../../APIKEY.md)。

## 调用关系

- 实现：[`internal/dao/entdao`](./entdao/db.md)
- 注入：`cmd/server/main.go` 构造实现后通过 `router.New` 注入（见 [`cmd/server.md`](../../cmd/server.md)、[`router.md`](../router/router.md)）。
