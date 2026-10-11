# `internal/dao/entdao/audio.go`

[`dao.AudioDAO`](../dao.md) 的 Ent + SQLite 实现，与 [image.go](image.md) 同构。编译期断言 `var _ dao.AudioDAO = (*audioDAO)(nil)`。

## 方法

| 方法 | 说明 |
| --- | --- |
| `Create` | 落库全部元数据字段（`SetNillableKeyID` 处理可空来源密钥），回填 ID/CreatedAt。 |
| `GetByID` / `GetByHash` | 按 PK / 内容哈希查询（秒传）；未找到 `wrapErr` → `dao.ErrNotFound`。 |
| `List` | `countAudios`（与 List 同过滤）+ 排序/分页查询；`KeyIDEQ` 过滤、`ent.Asc/Desc(FieldCreatedAt)` 排序（非 "desc" 一律升序）、`Offset/Limit` 分页。 |
| `ListByIDs` / `DeleteByIDs` | 批量删除用；空切片防御（不发起查询直接返回空/0）；不存在 ID 静默跳过。 |
| `ListByKeyID` / `DeleteByKeyID` | 删除 API 密钥时的级联清理（service/apikey.go 调用）。 |
| `UpdateCoverURL(id, coverURL *string)` | 改封面：**nil 走 `ClearCoverURL()`、非 nil 走 `SetCoverURL`**——ent 的 `SetNillableCoverURL(nil)` 是「不操作」而非「清空」，nil 分支必须显式 Clear 才能回到 NULL。返回更新后的完整记录；未找到 `dao.ErrNotFound`。 |

`audioToModel` 完成 `ent.Audio → model.Audio` 的全字段转换。

## 与其它文件的关系

- 接口定义：[`internal/dao/dao.go`](../dao.md) 的 `AudioDAO`；
- schema：[`ent/schema/audio.go`](../../../ent/schema/audio.md)；
- 消费方：`service.AudioService`、`service.APIKeyService.Delete`（级联）；
- 无仪表盘聚合方法（`Count`/`TotalSize`/`CountByRange*`）——仪表盘 v1 刻意只统计图片（见 `docs/backend/AUDIO.md` 的范围说明）。
