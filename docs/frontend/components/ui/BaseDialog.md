# `frontend/app/components/ui/BaseDialog.vue`

可复用的弹窗外壳，Nuxt 自动导入标签为 `<UiBaseDialog />`。供 APIkey 管理页的各个操作弹窗复用，统一遮罩 / ESC 关闭 / 头尾布局。

## 职责

- `<Teleport to="body">` 渲染到文档末尾，避免被父级 `overflow` 裁切。
- 全屏遮罩点击关闭、ESC 关闭、右上角关闭按钮。
- 提供 `header`（默认显示 `title`）、默认、`footer` 三个插槽。

## Props / Emits

- props：`open: boolean`（显隐）、`title?: string`（无 `header` 插槽时显示）、`contentClass?: string`（透传给内容容器，调整宽度）。
- emits：`close()`。

## 实现要点

- `watch(open)` 在显示 / 隐藏时增删 `window` 的 `keydown` ESC 监听，`onBeforeUnmount` 兜底移除（模式同 [`content/ImageDetailDialog`](../content/image-detail-dialog.md)）。
- `v-if="open"` 控制挂载，关闭即卸载。

## 与其它文件的关系

- 被 `components/apikeys/*` 各弹窗作为外壳使用。
