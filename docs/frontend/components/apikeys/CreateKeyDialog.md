# `frontend/app/components/apikeys/CreateKeyDialog.vue`

创建 API Key 的表单弹窗，Nuxt 自动导入标签为 `<ApikeysCreateKeyDialog />`。

## 职责

- 输入密钥名称 + 选择权限范围（读写 / 只读，默认读写）。
- 提交调 [`useApiKeys`](../../composables/useApiKeys.md).`create`，成功后 emit `created` 交由父组件展示一次性明文。

## Props / Emits

- props：`open: boolean`。
- emits：`close()`、`created(resp: CreateAPIKeyResponse)`。

## 实现要点

- 表单用 `reactive` + `validate`（仿 [`login/LoginForm`](../login/LoginForm.md)）：名称必填。
- `watch(open)` 每次打开重置表单与错误态。
- `rate_limit` 不暴露（沿用全局默认）；`scope` 用两宫格按钮选择。

## 与其它文件的关系

- 父组件：[`pages/apikeys/index.vue`](../../pages/apikeys/index.md)。
- 外壳：[`ui/BaseDialog`](../ui/BaseDialog.md)。接口：[`useApiKeys`](../../composables/useApiKeys.md)。
