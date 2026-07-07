# `frontend/app/components/apikeys/ResetKeyDialog.vue`

重置 API Key 明文的确认弹窗，Nuxt 自动导入标签为 `<ApikeysResetKeyDialog />`。

## 职责

- 说明重置后果（旧明文失效、取消吊销、新明文仅显示一次），确认后调 [`useApiKeys`](../../composables/useApiKeys.md).`reset`，成功 emit `reset(resp)` 交由父组件展示新明文。

## Props / Emits

- props：`open: boolean`、`apiKey: APIKeyInfo | null`。
- emits：`close()`、`reset(resp: ResetAPIKeyResponse)`。

## 与其它文件的关系

- 父组件：[`pages/apikeys/index.vue`](../../pages/apikeys/index.md)。
- 外壳：[`ui/BaseDialog`](../ui/BaseDialog.md)。
