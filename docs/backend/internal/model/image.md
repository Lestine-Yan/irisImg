# internal/model/image.go

图片元信息的**跨层数据载体**（实体 DTO）。

## 类型：Image

独立于 Ent 生成的 `ent.Image`：DAO 层负责在二者之间转换（见 [`entdao/image.go`](../dao/entdao/image.md) 的 `toModel`），使 service / api 层不直接依赖 Ent，便于替换存储实现。

字段：`ID`、`Filename`、`StoredPath`、`URL`、`Size`、`MimeType`、`Width`、`Height`、`Hash`、`CreatedAt`，均带 `json` 标签，可直接作为 API 响应体。字段语义与 [`ent/schema/image.go`](../../ent/schema/image.md) 一一对应。
