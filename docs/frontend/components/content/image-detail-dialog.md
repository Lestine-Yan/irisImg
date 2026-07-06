# `frontend/app/components/content/ImageDetailDialog.vue`

内容中心点击图片后弹出的详情数据框。Nuxt 自动导入标签为 `<ContentImageDetailDialog />`。

## 职责

- 全屏遮罩 + 居中卡片：左侧大图预览，右侧元数据列表（来源 Key、尺寸、大小、类型、上传时间、文件 ID、哈希、访问 URL）。
- 来源 Key 名称由父组件传入（前端 `id → name` 映射解析，无需后端联表）。
- 点击遮罩 / 关闭按钮 / 按 ESC 关闭，emit `close`。

## Props / Emits

- props：`image: ImageItem | null`（null 时不渲染）、`keyName?: string`。
- emits：`close()`。

## 实现要点

- 用 `<Teleport to="body">` 渲染到文档末尾，避免被父级 `overflow` 裁切。
- `watch(image)` 在弹窗显示 / 隐藏时增删 `window` 的 `keydown` ESC 监听，`onBeforeUnmount` 兜底移除，避免泄漏。

## 与其它文件的关系

- 父组件：[`pages/content/index.vue`](../../pages/content/index.md)。
- 依赖 [`useImages`](../../composables/useImages.md) 的 `resolveImageUrl` / `formatBytes` / `formatDate`（自动导入）。
