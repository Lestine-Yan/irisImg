# `frontend/app/layouts/default.vue`

后台默认布局：左侧 `<LayoutAppSidebar>` + 右侧内容区。所有未显式声明 `layout: false` 的页面都走此布局。

## 职责

- 用 flex 容器组合侧边栏与主内容区。
- 主内容区提供 `max-w-6xl` 居中容器与统一内边距。
- 挂载后统一调用 `fetchMe()` 刷新用户信息，供侧边栏展示。

## 视觉

- 外层 `flex min-h-screen bg-iris-cream`（暖米白留白底）。
- 侧边栏 sticky 固定在左侧；主内容区 `flex-1` 滚动。

## 鉴权逻辑

- 布局本身不做登录态校验，由各页面 `definePageMeta({ middleware: 'auth' })` 保护。
- `onMounted` 中若 `isAuthenticated` 为真则 `fetchMe()`；token 仅客户端可见，故仅在客户端触发。

## 与其它文件的关系

```
default.vue
  ├── AppSidebar.vue
  │     └── UserBadge.vue
  └── useAuth.ts
        └── fetchMe() → /api/v1/auth/me
```

## 修改建议

- 若所有后台页都需要鉴权，可考虑改为全局 middleware 统一保护，省去每页 `definePageMeta`。
- 需要顶部栏（面包屑、全局搜索）时可在主内容区上方加 header 行。
