# `frontend/app/components/apikeys/RevokeDeleteDialog.vue`

吊销 / 删除 API Key 的危险操作弹窗，Nuxt 自动导入标签为 `<ApikeysRevokeDeleteDialog />`。

## 职责

- 单选操作模式：「仅吊销」（软删除）或「删除并清理图片」（硬删 + 级联删图）。
- 各模式独立的警告文本。
- 账号密码二次确认表单（`username` 预填当前登录用户，可改；`password`）。
- 提交按模式调 [`useApiKeys`](../../composables/useApiKeys.md).`revoke` 或 `purge`，成功 emit `done(mode)`。

## Props / Emits

- props：`open: boolean`、`apiKey: APIKeyInfo | null`。
- emits：`close()`、`done(mode: 'revoke' | 'purge')`。

## 实现要点

- `canSubmit` 计算属性：用户名与密码均非空才允许提交。
- 删除模式按钮用 `rose-600` 危险色，吊销模式用 `iris-dark`。
- 密码错误时后端返回 403，`useApi` 不登出，本组件就地显示「用户名或密码错误」。

## 与其它文件的关系

- 父组件：[`pages/apikeys/index.vue`](../../pages/apikeys/index.md)。
- 外壳：[`ui/BaseDialog`](../ui/BaseDialog.md)。当前用户名取自 [`useAuth`](../../composables/useAuth.md)。
