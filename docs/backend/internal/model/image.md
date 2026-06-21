# internal/model/image.go

图片元信息的**跨层数据载体**（实体 DTO）。

## 类型：Image

独立于 Ent 生成的 `ent.Image`：DAO 层负责在二者之间转换（见 [`entdao/image.go`](../dao/entdao/image.md) 的 `toModel`），使 service / api 层不直接依赖 Ent，便于替换存储实现。

字段：`ID`、`Filename`、`StoredPath`、`URL`、`Size`、`MimeType`、`Width`、`Height`、`Hash`、`CreatedAt`，均带 `json` 标签，可直接作为 API 响应体。字段语义与 [`ent/schema/image.go`](../../ent/schema/image.md) 一一对应。

此外含一个可空字段：

- `KeyID *int`（`json:"key_id,omitempty"`）：添加该图片的 API 密钥 ID。通过后台 JWT 上传的图片没有关联密钥，此处为 `nil`（序列化时省略）；通过密钥 POST 添加的图片回填中间件注入的 `api_key_id`。对应 schema 的 `key` edge / `key_id` 字段，详见 [`APIKEY.md`](../../APIKEY.md)。
