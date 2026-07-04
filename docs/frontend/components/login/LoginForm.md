# `frontend/app/components/login/LoginForm.vue`

登录页右侧白色卡片登录表单组件。

## 职责

- 提供用户名、密码输入框。
- 执行前端校验：
  - 用户名非空。
  - 密码非空且长度不少于 4 位。
- 提交时调用 `useAuth().login()`。
- 登录成功后 `emit('success')`。
- 展示校验错误与后端返回的错误信息。

## Props / Emits

- 无 props。
- `success`：登录成功时触发，由父页面跳转。

## 状态

- `form`：用户名、密码输入值。
- `errors`：按字段的校验错误。
- `serverError`：后端返回的错误信息。
- `loading`：提交中的 loading 状态，用于禁用按钮并显示 spinner。

## 视觉

- 卡片为白底圆角；输入框聚焦态用 `iris-dark`（边框 + ring）。
- 「进入工作台」按钮为 `iris-dark` 底 + 白字，hover 底色转 `iris-violet`；loading spinner 为白色。
- 已移除底部「还没有账号？立即注册」段落。

## 与其它文件的关系

```
LoginForm.vue
  └── useAuth.ts
        └── login(username, password) → useApi.ts → POST /auth/login
```

## 修改建议

- 若后续需要恢复注册入口，可重新加回底部链接或跳转按钮。
- 若引入表单校验库（如 vee-validate / zod），可替换当前的轻量 `reactive` 校验逻辑。
