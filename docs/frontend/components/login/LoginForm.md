# `frontend/app/components/login/LoginForm.vue`

登录页右侧半透明登录卡片组件。

## 职责

- 提供用户名、密码输入框。
- 执行前端校验：
  - 用户名非空。
  - 密码非空且长度不少于 4 位。
- 提交时调用 `useAuth().login()`。
- 登录成功后 `emit('success')`。
- 展示字段校验错误与后端返回的错误信息。

## Props / Emits

- 无 props。
- `success`：登录成功时触发，由父页面跳转。

## 状态

- `form`：用户名、密码输入值。
- `errors`：按字段的校验错误。
- `serverError`：后端返回的错误信息。
- `loading`：提交中的 loading 状态，用于禁用按钮并显示 spinner。
- `showPassword`：密码可见性开关，配合 `passwordType` 计算属性切换 `text/password`。

## 视觉

- 卡片为白色半透明（`bg-white/90`）、26px 圆角、白色细边框、柔和多层阴影与 `backdrop-blur-xl`；外层另有一圈低透明度渐变光晕。
- 顶部 `SIGN IN` 标签 + 绿色圆点的「私有部署」胶囊。
- 输入框为 48px 高、圆角、浅灰边框；聚焦时紫罗兰边框 + 4px 低透明度 ring。
- 字段标签明确可见，右侧配 `USERNAME / PASSWORD` 微型英文标签。
- 密码框内置显示/隐藏按钮，带动态 `aria-label` 与 `aria-pressed`。
- 主按钮为近黑紫底 + 白色文字，hover 转 `iris-dark`，右侧箭头随 hover 微移；loading 时显示 spinner。
- 底部用细分割线隔开「凭据仅提交至你的服务器」与 `Enter` 提示。
- 表单校验错误与后端错误均带警示图标；`serverError` 使用 `role="alert"`。
- 卡片带 `card-in` 入场动画；`prefers-reduced-motion` 下关闭。

## 可访问性

- 用户名输入：`autocomplete="username"`、`autocapitalize="off"`、`spellcheck="false"`。
- 密码输入：`autocomplete="current-password"`。
- 错误字段通过 `aria-invalid` 与 `aria-describedby` 关联到对应错误文本。

## 与其它文件的关系

```
LoginForm.vue
  └── useAuth.ts
        └── login(username, password) → useApi.ts → POST /auth/login
```

登录链路本身未变化，详见 `docs/frontend/AUTH.md`。

## 修改建议

- 若后续需要恢复注册入口或「忘记密码」，可在底部提示区扩展。
- 若引入表单校验库（如 vee-validate / zod），可替换当前的轻量 `reactive` 校验逻辑。
