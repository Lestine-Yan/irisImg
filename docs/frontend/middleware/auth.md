# `frontend/app/middleware/auth.ts`

客户端命名路由守卫，保护后台页面仅对已登录用户开放。

## 职责

- 在进入挂载了 `middleware: 'auth'` 的页面前校验登录态。
- 未登录时重定向到登录页 `/`。

## 运行时机

- 项目为 SPA 模式（`nuxt.config.ts` 中 `ssr: false`），无服务端渲染，middleware 仅在浏览器执行。
- 客户端插件 `plugins/auth.client.ts` 先于 middleware 执行，token 已在启动时从 `localStorage` 恢复，运行到此处时登录态已就绪。

## 与其它文件的关系

```
auth.ts
  └── useAuth.ts
        └── isAuthenticated（基于 token + expiresAt 计算）
```

## 使用方式

页面中通过 `definePageMeta` 引用：

```ts
definePageMeta({ middleware: 'auth' })
```

当前使用该守卫的页面：`pages/dashboard.vue`、`pages/content/index.vue`、`pages/logs/index.vue`、`pages/apikeys/index.vue`、`pages/settings/index.vue`。

## 修改建议

- 若后续改为 SSR 模式并希望服务端也校验，需先把鉴权迁移到 cookie + httpOnly，让服务端能读到登录态。
- 也可改为全局 middleware（`middleware/auth.global.ts`）统一保护 `/` 之外的所有路由，省去每页 `definePageMeta`。
