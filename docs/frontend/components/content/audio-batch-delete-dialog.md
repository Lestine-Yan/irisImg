# `frontend/app/components/content/AudioBatchDeleteDialog.vue`

音频批量删除确认弹窗，Nuxt 自动导入标签为 `<ContentAudioBatchDeleteDialog />`。与 [`BatchDeleteDialog.vue`](./batch-delete-dialog.md) 逐结构镜像（基于 [`UiBaseDialog`](../ui/BaseDialog.md) 的标题 / 警告块 / 账号密码二次确认 / 内联错误 / 底部按钮组），仅文案与接口不同：

| 维度 | BatchDeleteDialog（图片） | AudioBatchDeleteDialog（音频） |
| --- | --- | --- |
| 标题 / 警告 | 「批量删除图片」/「N 张图片」 | 「批量删除音频」/「N 个音频」 |
| 接口 | `useImages().batchRemove` → `DELETE /admin/images` | `useAudios().batchRemove` → `DELETE /admin/audios` |
| 响应类型 | `BatchDeleteImagesResult` | `BatchDeleteAudiosResult` |

行为细节与图片版一致：打开时预填当前用户名（`useAuth().user`）并清空密码 / 错误；`canSubmit` 要求用户名 + 密码 + 非空 ids；后端二次确认失败返回 **403**（不触发 useApi 的全局登出），弹窗保持打开内联报错可重试。

## Props / Emits

- props：`open: boolean`；`ids: number[]`。
- emits：`close: []`；`done: [result: BatchDeleteAudiosResult]`。

## 与其它文件的关系

- 父组件：[`pages/content/audio.vue`](../../pages/content/audio.md)（多选删除入口，`onDeleted` 收敛页码并刷新）。
- 依赖 [`useAudios`](../../composables/useAudios.md)（`batchRemove`）、[`useAuth`](../../composables/useAuth.md)（预填用户名）。
