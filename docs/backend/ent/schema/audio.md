# `ent/schema/audio.go`

> 文档路径约定：`ent/schema/audio.go` → `docs/backend/ent/schema/audio.md`。

音频文件的 Ent 实体 schema，由 `go generate ./ent` 生成 `ent/audio*` 代码（生成树整体提交）。与 [image.go](image.md) 平行的存储字段 + 标准音频元数据列。

## 字段

**存储字段**（与 images 同构）：`filename`（原始文件名）、`stored_path`（相对音频根目录的落盘路径，**Unique**）、`url`、`size`（Int64，非负）、`mime_type`（规范化值）、`hash`（sha256，秒传/去重）、`created_at`（默认 now，Immutable）、`key_id`（Optional/Nillable，API 密钥来源）。

**格式**：`format`（mp3/m4a/flac，解析失败为空串）。

**标签元数据**（全部零值 = 未知/缺失，解析失败不阻断上传）：`title`、`artist`、`album`、`album_artist`、`composer`、`genre`、`year`、`track_number`、`disc_number`。

**技术参数**（零值 = 未知）：`duration_seconds`（Float）、`bitrate_kbps`（平均码率）、`sample_rate`（Hz）、`channels`、`bit_depth`（仅 FLAC 有意义，其余为 0）。

**双层封面**：

- `embedded_cover_url`（默认空串）——上传时从内嵌封面抽取落盘的 URL（`covers/<sha256>.<ext>` 内容寻址共享），此后不变，无内嵌封面为空；
- `cover_url`（Optional/Nillable）——用户自定义覆盖，`nil` = 未覆盖。生效封面 = `cover_url || embedded_cover_url`（前端计算）。

## Edges / 索引 / 外键

- `edge.From("key", ApiKey.Type).Ref("audios").Field("key_id").Unique()`：可选归属密钥（后台 JWT 直传为空）；生成外键 `audios_api_keys_audios` **ON DELETE SET NULL**（与图片一致——即便如此，`APIKeyService.Delete` 仍先级联删音频记录再删密钥，保证计数与文件清理）。
- 索引：`hash`（秒传查询）、`created_at`（分页列表），均非唯一；内容唯一性由 `stored_path` Unique + service 层 `GetByHash` 去重保证。

## 迁移

表 `audios` 由 `client.Schema.Create`（`database.auto_migrate: true`）自动建表——纯新增表，老库升级无破坏性。
