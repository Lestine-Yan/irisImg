# `frontend/app/pages/dashboard.vue`

工作台占位页，对应路由 `/dashboard`。

## 职责

- 登录成功后跳转目标。
- 客户端校验登录态，未登录则跳转 `/`。
- 展示当前登录用户名。
- 提供退出登录按钮。

## 鉴权逻辑

- `onMounted` 中检查 `isAuthenticated.value`：
  - 未登录 → `navigateTo('/')`。
  - 已登录 → 调用 `useAuth().fetchMe()` 刷新用户信息。

## 与其它文件的关系

```
dashboard.vue
  └── useAuth.ts
        ├── isAuthenticated
        ├── fetchMe() → /api/v1/auth/me
        └── logout() → navigateTo('/')
```

## 修改建议

- 当前为占位页，后续可替换为真正的图片上传/管理功能。
- 若页面数量增多，建议抽取 `middleware/auth.ts` 统一做客户端路由守卫，减少重复校验代码。
