# internal/dao/entdao/image.go

[`dao.ImageDAO`](../dao.md) 的 Ent 实现。

## 类型

- `imageDAO`：持有 `*ent.Client` 的非导出实现类型。
- `NewImageDAO(client *ent.Client) dao.ImageDAO`：构造函数。
- 编译期断言 `var _ dao.ImageDAO = (*imageDAO)(nil)` 保证接口一致。

## 方法

逐一实现 `ImageDAO` 接口：`Create` / `GetByID` / `GetByHash` / `List` / `Delete`。

- `Create`：`SetNillableKeyID(img.KeyID)` 写入可空外键 `key_id`（记录图片由哪把 API 密钥添加；JWT 上传时为 nil 则不设置）。
- `List(ctx, q model.ImageListQuery)`：按 `q.KeyID` 过滤（非 nil 时用 `image.KeyIDEQ`）、按 `q.Order` 排序（`"desc"` 倒序，否则升序）、`q.Offset`/`q.Limit` 为正才生效；总数与过滤条件一致，由私有 `countImages` 统计。
- 查询单条用 ent 生成的 `image.HashEQ` 等谓词。

## 错误与转换

- `wrapErr`：`ent.IsNotFound(err)` → [`dao.ErrNotFound`](../errors.md)，屏蔽 Ent 错误类型。本函数同时被同包 [`apikey.go`](apikey.md) 复用。
- `toModel(*ent.Image) *model.Image`：把 Ent 实体转换为跨层的 [`model.Image`](../../model/image.md)，并回填 `KeyID`，使 service / api 不直接依赖 Ent。

## 测试

`image_test.go` 在 `t.TempDir()` 打开真实 SQLite（纯 Go 驱动、离线、无 CGO），覆盖创建/查询、未找到、列表分页与删除；`TestImageDAO_ListFilterAndOrder` 额外覆盖按 `key_id` 过滤、asc/desc 排序、offset/limit 分页（用 ent client 写入带明确 `created_at` 的记录以稳定排序，并预置 `api_key` 行满足外键约束）。
