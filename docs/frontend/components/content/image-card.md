# `frontend/app/components/content/ImageCard.vue`

内容中心图片网格中的单张卡片。Nuxt 自动导入标签为 `<ContentImageCard />`。

## 职责

- 展示缩略图（`resolveImageUrl(image.url)` + `object-cover` 正方形）、文件名、大小与上传时间。
- 点击 emit `click(image)`，由父页面打开详情弹窗。

## Props / Emits

- props：`image: ImageItem`（类型见 [`useImages`](../../composables/useImages.md)）。
- emits：`click(image: ImageItem)`。

## 与其它文件的关系

- 父组件：[`pages/content/index.vue`](../../pages/content/index.md)。
- 依赖 [`useImages`](../../composables/useImages.md) 的 `resolveImageUrl` / `formatBytes` / `formatDate`（自动导入）。
