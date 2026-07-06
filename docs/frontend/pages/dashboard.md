# `frontend/app/pages/dashboard.vue`

仪表盘占位页，对应路由 `/dashboard`，使用默认布局 `layouts/default.vue`。

## 职责

- 登录成功后跳转目标，也是后台默认着陆页。
- 暂只展示标题、副标题与「正在开发中」占位卡片。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`：未登录跳转 `/`。
- 校验已集中到 `middleware/auth.ts`，页面内不再做 `onMounted` 重复校验。
- 用户信息刷新由 `layouts/default.vue` 的 `onMounted` 统一调用 `fetchMe()` 完成。

## 与其它文件的关系

```
dashboard.vue
  └── middleware/auth.ts
        └── useAuth.ts → isAuthenticated
layouts/default.vue
  ├── AppSidebar.vue
  └── useAuth.ts → fetchMe()
```

## 修改建议

- 后续替换为真正的运营概览卡片（存储用量、最近上传、API 调用统计等）。
- 若需要页面级数据预取，可改用 `useAsyncData` 在 SSR 阶段拉取。
