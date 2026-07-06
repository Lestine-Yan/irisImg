# `frontend/app/composables/useImages.ts`

图片资源请求封装 + 类型定义 + 图片 URL / 体积 / 时间格式化工具，供内容中心使用。

## 导出类型

- `ImageItem`：单张图片元信息，对应后端 `model.Image`（`id / filename / stored_path / url / size / mime_type / width / height / hash / created_at / key_id`）。
- `ImageListResponse`：`GET /admin/images` 的响应 `data`，含 `items / total / page / page_size`。
- `ListImagesParams`：`list()` 的入参（`keyId / order / page / pageSize`）。

## 导出函数

### `useImages()`

返回：

- `list(params)`：调 [`useApi`](./useApi.md) 的 `get('/admin/images', { query })`，自动附带 JWT。`keyId` 不传则不按密钥过滤；`order` 默认 `asc`（时间升序）。
- `upload(file)`：调 [`useApi`](./useApi.md) 的 `post('/admin/images', FormData)`，走后台 JWT 直传通道（`POST /admin/images`），自动附带 JWT。`FormData` 字段名固定 `file`；ofetch 自动设置 `multipart/form-data` 边界，**不要手动设 Content-Type**。返回新增的 `ImageItem`（`key_id` 为 `null`，即 admin 直传，不关联任何密钥）。与对外 `POST /images`（API Key 鉴权）解耦。

### `resolveImageUrl(url)`

把后端返回的图片 URL 解析成浏览器可加载的完整地址：

- 后端 `storage.public_base_url` 为空时，`url` 形如 `/imgs/2026/07/<hash>.png`（相对路径）。前端 SPA 跑在另一端口（如 :3000），需拼上后端 origin（从 `runtimeConfig.public.apiBase` 去掉 `/api/v1` 推导）。
- `url` 已是 `http(s)://` 绝对地址时原样返回。

### `formatBytes(n)` / `formatDate(s)`

体积（B / KB / MB / GB）与时间（`YYYY-MM-DD HH:mm` 本地时间）格式化，供卡片与详情弹窗复用。

## 与其它文件的关系

- 被 [`pages/content/index.vue`](../pages/content/index.md) 与 [`components/content/ImageCard.vue`](../components/content/image-card.md)、[`ImageDetailDialog.vue`](../components/content/image-detail-dialog.md)、[`UploadPanel.vue`](../components/content/upload-panel.md) 使用。
- 底层依赖 [`useApi.ts`](./useApi.md)。
