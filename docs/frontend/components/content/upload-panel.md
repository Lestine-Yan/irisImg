# `frontend/app/components/content/UploadPanel.vue`

内容中心的图片上传栏。Nuxt 自动导入标签为 `<ContentUploadPanel />`。

## 职责

- 提供拖拽区 + 点击选择（`<input type="file" accept="image/*" multiple>`），多选可批量上传。
- 维护上传队列，逐项展示文件名、大小（`formatBytes`）与状态（等待中 / 上传中 / 已完成 / 错误），错误信息取后端返回的业务 `message`。
- 顺序上传（一次一张），每张成功即 `emit('uploaded', img)`，由父页面刷新列表；提供「移除」「清空已完成」操作。
- 拖拽高亮用深度计数（`dragDepth`）避免子元素反复触发 `dragenter`/`dragleave` 闪烁；`dragover` `preventDefault` 以保证 `drop` 触发。

## Props / Emits

- props：无。
- emits：`uploaded(img: ImageItem)`——每张上传成功后触发，`img` 为新增图片（`key_id` 为 `null`，即 admin 直传）。

## 鉴权与数据流

- 调 [`useImages`](../../composables/useImages.md) 的 `upload(file)`，走 `POST /admin/images`（JWT 通道，由 [`useApi`](../../composables/useImages.md) 自动附带 `Authorization`）。
- 上传的图片**不关联密钥**（`key_id` 留空），只会在内容中心「全部」里出现，详情里来源展示为 `admin`。
- 错误信息：ofetch 把非 2xx 响应体挂在 `err.data`，`resolveErrorMsg` 优先取 `err.data.message`（如「上传文件过大」「不支持的图片类型」），兜底 `err.message` / 「上传失败」。

## 与其它文件的关系

- 父组件：[`pages/content/index.vue`](../../pages/content/index.md)（标题右侧「上传图片」按钮切换其显隐，`@uploaded` 触发列表刷新）。
- 依赖 [`useImages`](../../composables/useImages.md) 的 `upload` / `formatBytes`（自动导入）。
