# internal/model/image.go

图片元信息的**跨层数据载体**（实体 DTO）。

## 类型：Image

独立于 Ent 生成的 `ent.Image`：DAO 层负责在二者之间转换（见 [`entdao/image.go`](../dao/entdao/image.md) 的 `toModel`），使 service / api 层不直接依赖 Ent，便于替换存储实现。

字段：`ID`、`Filename`、`StoredPath`、`URL`、`Size`、`MimeType`、`Width`、`Height`、`Hash`、`CreatedAt`，均带 `json` 标签，可直接作为 API 响应体。字段语义与 [`ent/schema/image.go`](../../ent/schema/image.md) 一一对应。

此外含一个可空字段：

- `KeyID *int`（`json:"key_id,omitempty"`）：添加该图片的 API 密钥 ID。通过后台 JWT 上传的图片没有关联密钥，此处为 `nil`（序列化时省略）；通过密钥 POST 添加的图片回填中间件注入的 `api_key_id`。对应 schema 的 `key` edge / `key_id` 字段，详见 [`APIKEY.md`](../../APIKEY.md)。

## 类型：UploadImageInput

「上传一张图片」的入参，由 api 层从 HTTP 请求装配后传给 [`service.ImageService.Upload`](../service/image.md)。设计成结构体（而非多参数）便于后续追加字段（如标签、相册 ID）不破坏签名。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `Filename` | string | 客户端给出的原始文件名，仅作展示。真实落盘文件名由 hash + 嗅探的扩展名决定 |
| `Content` | []byte | 完整字节，由 api 层在 `http.MaxBytesReader` 保护下读出 |
| `KeyID` | *int | 添加该图片的 API 密钥 ID。API Key 渠道由中间件保证非空；JWT 直传渠道（暂未实现）传 nil |
