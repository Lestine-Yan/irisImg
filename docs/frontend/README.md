# frontend 文档索引

本目录与 `frontend/app/` 保持一一对应：每个源文件 `frontend/app/<path>/<name>.<ext>` 对应一篇 `docs/frontend/<path>/<name>.md`，描述其职责、对外行为与调用关系。`frontend/` 根下的配置文件（`nuxt.config.ts`、`tailwind.config.ts`）同样在此对照。

## 目录结构对照

```
frontend/                                  docs/frontend/
├── nuxt.config.ts                          ├── nuxt.config.md
├── tailwind.config.ts                      ├── tailwind.config.md
└── app/
    ├── app.vue                             ├── app.md
    ├── assets/css/tailwind.css             ├── assets/css/tailwind.md
    ├── layouts/
    │   └── default.vue                     │   └── layouts/default.md
    ├── pages/
    │   ├── index.vue                       │   ├── pages/index.md
    │   ├── dashboard.vue                   │   ├── pages/dashboard.md
    │   ├── content/index.vue               │   ├── pages/content/index.md
    │   ├── logs/index.vue                  │   ├── pages/logs/index.md
    │   ├── apikeys/index.vue               │   ├── pages/apikeys/index.md
    │   └── settings/index.vue              │   └── pages/settings/index.md
    ├── components/
    │   ├── login/
    │   │   ├── GeometricBackground.vue     │   │   ├── components/login/GeometricBackground.md
    │   │   ├── LoginHero.vue               │   │   ├── components/login/LoginHero.md
    │   │   └── LoginForm.vue               │   │   └── components/login/LoginForm.md
    │   └── layout/
    │       ├── AppSidebar.vue              │       ├── components/layout/AppSidebar.md
    │       └── UserBadge.vue               │       └── components/layout/UserBadge.md
    ├── composables/
    │   ├── useAuth.ts                      │   ├── composables/useAuth.md
    │   └── useApi.ts                       │   └── composables/useApi.md
    ├── middleware/
    │   └── auth.ts                         │   └── middleware/auth.md
    └── plugins/
        └── auth.client.ts                  └── plugins/auth.client.md
```

## 跨文件特性文档

- [AUTH.md](./AUTH.md) — 前端登录链路：页面 → 表单 → useAuth → useApi → 后端 `/auth/login`、`/auth/me`。

## 分层说明

- **根配置**：`nuxt.config.ts` 为 Nuxt 入口（SPA 模式 `ssr: false`、模块、runtimeConfig）；`tailwind.config.ts` 定义 iris 全局色系；`assets/css/tailwind.css` 引入 Tailwind 三层。
- **layouts/**：Nuxt 布局。`default.vue` 是后台默认布局，左侧 `AppSidebar` + 右侧内容区，并在挂载后统一 `fetchMe()` 刷新用户信息。
- **pages/**：Nuxt 文件路由，每个 `.vue` 对应一个 URL。`index.vue` 是登录落地页（`layout: false`）；`dashboard.vue`、`content/`、`logs/`、`apikeys/`、`settings/` 是后台五个模块的占位页，均通过 `middleware: 'auth'` 保护。
- **components/login/**：登录页专用展示组件。`GeometricBackground.vue` 负责 SVG 背景；`LoginHero.vue` 负责左侧品牌字标；`LoginForm.vue` 负责右侧登录表单与校验。
- **components/layout/**：后台布局专用组件。`AppSidebar.vue` 负责侧边栏（logo + 导航 + 用户区）；`UserBadge.vue` 负责底部当前用户展示与退出登录。
- **composables/**：自动导入的组合函数。`useAuth.ts` 维护 token 与登录态；`useApi.ts` 封装 `$fetch`，统一处理响应体、鉴权头与 401 跳转。
- **middleware/**：路由中间件。`auth.ts` 是客户端命名守卫，未登录访问后台页面时跳转 `/`。
- **plugins/**：客户端插件。`auth.client.ts` 在应用启动时从 `localStorage` 恢复 token。

## 入口阅读顺序

1. [nuxt.config.md](./nuxt.config.md) — Nuxt 配置入口（SPA 模式、API 基址）。
2. [app.md](./app.md) — 应用根组件。
3. [layouts/default.md](./layouts/default.md) — 后台布局骨架。
4. [components/layout/AppSidebar.md](./components/layout/AppSidebar.md) — 侧边栏与导航。
5. [AUTH.md](./AUTH.md) — 登录链路总览。
6. [middleware/auth.md](./middleware/auth.md) — 后台路由守卫。
7. 按需阅读具体页面与组件文档。

## 环境变量

- `NUXT_PUBLIC_API_BASE`：前端请求后端的基址，默认 `http://localhost:8080/api/v1`。
