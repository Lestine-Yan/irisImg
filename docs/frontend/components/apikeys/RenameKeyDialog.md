# `frontend/app/components/apikeys/RenameKeyDialog.vue`

重命名 API Key 的表单弹窗，Nuxt 自动导入标签为 `<ApikeysRenameKeyDialog />`。

## 职责

- 输入新名称，提交调 [`useApiKeys`](../../composables/useApiKeys.md).`rename`，成功 emit `renamed`。

## Props / Emits

- props：`open: boolean`、`apiKey: APIKeyInfo | null`。
- emits：`close()`、`renamed()`。

## 实现要点

- `watch(open)` 打开时用 `apiKey.name` 预填。
- 名称必填校验。

## 与其它文件的关系

- 父组件：[`pages/apikeys/index.vue`](../../pages/apikeys/index.md)。
- 外壳：[`ui/BaseDialog`](../ui/BaseDialog.md)。
