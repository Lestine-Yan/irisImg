# `frontend/app/composables/useAudios.ts`

封装音频中心用到的列表请求、后台直传上传、批量删除与改封面，外加封面 / 时长格式化工具。接口契约见后端 [`AUDIO.md`](../../backend/AUDIO.md)。

## 类型

- `AudioItem`：单个音频元信息（对应后端 `model.Audio`）。存储字段与 `ImageItem` 同构，外加 `format`、标签字段（`title`/`artist`/`album`/`album_artist`/`composer`/`genre`/`year`/`track_number`/`disc_number`）、技术参数（`duration_seconds`/`bitrate_kbps`/`sample_rate`/`channels`/`bit_depth`，未知为 0）与**双层封面**：`embedded_cover_url: string`（上传时抽取，此后不变）与 `cover_url?: string | null`（用户覆盖，未覆盖为空）。
- `AudioListResponse` / `ListAudiosParams`：与 `ImageListResponse` / `ListImagesParams` 同构（`keyId`/`order`/`page`/`pageSize`）。
- `BatchDeleteAudiosRequest` / `BatchDeleteAudiosResult`：批量删除的请求（`username`/`password`/`ids`）与「实际删除」口径响应（`deleted`/`ids`）。

## `useAudios()` 接口

| 方法 | 路由 | 说明 |
| --- | --- | --- |
| `list(params)` | `GET /admin/audios` | JWT 通道；query 同图片列表，返回 `{items, total, page, page_size}` |
| `upload(file)` | `POST /admin/audios` | FormData 字段 `file`，JWT 后台直传（`key_id` 为 null）；后端按内容魔数校验，扩展名伪造无效 |
| `batchRemove(params)` | `DELETE /admin/audios` | 账号密码二次确认（失败 403，不触发全局登出，弹窗内联报错） |
| `updateCover(id, coverUrl)` | `PATCH /admin/audios/:id` | body `{cover_url}`；**空串 = 清除覆盖**（恢复内嵌封面）；返回更新后的 `AudioItem`，调用方就地替换列表项 |

## 导出的工具函数

- `effectiveCover(item)`：**生效封面 = `cover_url || embedded_cover_url || ''`**（空串时 UI 显示占位音符）。
- `resolveAudioUrl(url)`：相对路径 `/audios/...` 拼后端 origin——实现上复用 [`resolveImageUrl`](./useImages.md)（其逻辑对任意相对 URL 通用），绝对地址原样返回。
- `formatDuration(seconds)`：`213.4 → "3:33"`、`≥1h → "1:02:03"`；0 / 空 / 非法值返回 `"—"`（后端未知时 `duration_seconds` 为 0）。

## 与其它文件的关系

- 消费方：[`pages/content/audio.vue`](../pages/content/audio.md)（列表/上传/批删/详情接线）、[`AudioCard.vue`](../components/content/audio-card.md)、[`AudioDetailDialog.vue`](../components/content/audio-detail-dialog.md)、[`AudioUploadPanel.vue`](../components/content/audio-upload-panel.md)、[`AudioBatchDeleteDialog.vue`](../components/content/audio-batch-delete-dialog.md)。
- 依赖 [`useApi`](./useApi.md)（JWT 头 + 响应解包 + 401 全局登出）。
