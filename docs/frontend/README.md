# frontend 文档索引

本目录与 `frontend/app/` 保持一一对应：每个源文件 `frontend/app/<path>/<name>.<ext>` 对应一篇 `docs/frontend/<path>/<name>.md`，描述其职责、对外行为与调用关系。

## 目录结构对照

```
frontend/app/                              docs/frontend/
├── app.vue                                 ├── app.md
├── assets/css/tailwind.css                 ├── assets/css/tailwind.md
├── tailwind.config.ts                      ├── tailwind.config.md
├── pages/
│   ├── index.vue                           │   ├── pages/index.md
│   └── dashboard.vue                       │   └── pages/dashboard.md
├── components/login/
│   ├── GeometricBackground.vue             │   ├── components/login/GeometricBackground.md
│   ├── LoginHero.vue                       │   ├── components/login/LoginHero.md
│   └── LoginForm.vue                       │   └── components/login/LoginForm.md
├── composables/
│   ├── useAuth.ts                          │   ├── composables/useAuth.md
│   └── useApi.ts                           │   └── composables/useApi.md
├── plugins/
│   └── auth.client.ts                      │   └── plugins/auth.client.md
```

## 跨文件特性文档

- [AUTH.md](./AUTH.md) — 前端登录链路：页面 → 表单 → useAuth → useApi → 后端 `/auth/login`、`/auth/me`。

## 分层说明

- **pages/**：Nuxt 文件路由，每个 `.vue` 对应一个 URL。`index.vue` 是登录落地页，`dashboard.vue` 是登录后的占位工作台。
- **components/login/**：登录页专用的展示组件。`GeometricBackground.vue` 负责 SVG 背景；`LoginHero.vue` 负责左侧品牌文案与功能标签；`LoginForm.vue` 负责右侧登录表单与校验。
- **composables/**：自动导入的组合函数。`useAuth.ts` 维护 token 与登录态；`useApi.ts` 封装 `$fetch`，统一处理响应体、鉴权头与 401 跳转。
- **plugins/**：客户端插件。`auth.client.ts` 在应用启动时从 `localStorage` 恢复 token。

## 入口阅读顺序

1. [app.md](./app.md) — 应用根组件。
2. [AUTH.md](./AUTH.md) — 登录链路总览。
3. [pages/index.md](./pages/index.md) — 登录页。
4. [composables/useAuth.md](./composables/useAuth.md) 与 [composables/useApi.md](./composables/useApi.md) — 状态与请求封装。
5. 按需阅读具体组件文档。

## 环境变量

- `NUXT_PUBLIC_API_BASE`：前端请求后端的基址，默认 `http://localhost:8080/api/v1`。
