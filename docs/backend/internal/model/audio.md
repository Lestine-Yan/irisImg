# `internal/model/audio.go`

> 文档路径约定：`internal/model/audio.go` → `docs/backend/internal/model/audio.md`。

音频元信息的跨层数据载体（实体）与请求/响应 DTO，与 [image.go](image.md) 平行。`Audio` 独立于 Ent 生成的 `ent.Audio`，DAO 层负责二者转换，service / api 层不直接依赖 Ent。

## 类型

### `Audio`

与 [image.go](image.go) 平行的存储字段（`ID` / `Filename` / `StoredPath` / `URL` / `Size` / `MimeType` / `Hash` / `CreatedAt` / `KeyID *int`）+：

- 标签元数据：`Title` / `Artist` / `Album` / `AlbumArtist` / `Composer` / `Genre` / `Year` / `TrackNumber` / `DiscNumber`；
- 技术参数：`DurationSeconds float64` / `BitrateKbps` / `SampleRate` / `Channels` / `BitDepth`；
- **双层封面**：
  - `EmbeddedCoverURL string`（json `embedded_cover_url`）——上传时抽取的内嵌封面 URL，此后不变；
  - `CoverURL *string`（json `cover_url,omitempty`）——用户覆盖，`nil` = 未覆盖；生效封面 = `CoverURL || EmbeddedCoverURL`，由前端计算展示。
- `Format string`——容器格式（mp3/m4a/flac）。

`KeyID` 语义同图片：API Key 渠道上传非空，后台 JWT 直传为 nil（展示为 admin）。

### `UploadAudioInput`

`{ Filename, Content []byte, KeyID *int }`——api 层从 multipart 装配后传给 service。

### `AudioListQuery` / `AudioListResult`

`{ KeyID *int, Order string, Offset, Limit int }`（非 "desc" 一律升序；Limit<=0 由 service 兜底 24）与 `{ Items, Total }`，与图片同名结构语义一致。

### `BatchDeleteAudiosRequest` / `BatchDeleteAudiosResponse`

批量删除请求/响应。请求体 binding 与图片一致：账号密码 `required` + `ids` `required,min=1,max=100,dive,gt=0`；响应 `{ Deleted, IDs }` 按「实际删除」口径（不存在的 ID 静默跳过）。

### `UpdateAudioCoverRequest`

`{ CoverURL string }`（binding `max=2048`）——详情页改封面的请求体。**空串 = 清除用户覆盖**（恢复展示内嵌封面），非空须 ≤ 2048 字符。
