# 前端登录链路

本文档描述 irisImg 前端从登录页到工作台的身份认证流程，以及 token 的存储、使用与失效处理。

## 涉及文件

- `frontend/app/pages/index.vue` — 登录落地页，组合背景、Hero、表单。
- `frontend/app/components/login/LoginForm.vue` — 登录表单组件，负责校验与提交。
- `frontend/app/composables/useAuth.ts` — 维护 token、用户信息、登录/登出/恢复。
- `frontend/app/composables/useApi.ts` — 封装 `$fetch`，自动附加 `Authorization` 头并处理 401。
- `frontend/app/plugins/auth.client.ts` — 客户端启动时恢复 token。
- `frontend/app/middleware/auth.ts` — 客户端路由守卫，未登录访问后台页时跳转 `/`。
- `frontend/app/layouts/default.vue` — 后台默认布局，挂载后统一 `fetchMe()` 刷新用户信息。
- `frontend/app/pages/dashboard.vue` — 登录成功后的占位页面（后台默认着陆页）。
- 后端契约见 `docs/backend/AUTH.md`。

## 登录流程

```
用户填写表单
    ↓
LoginForm.vue 本地校验（用户名/密码非空、密码长度 ≥4）
    ↓
调用 useAuth().login(username, password)
    ↓
useApi().post('/auth/login', { username, password })
    ↓
后端 POST /api/v1/auth/login 返回 { token, token_type, expires_at }
    ↓
useAuth 将 token 写入 useState 与 localStorage
    ↓
useAuth 调用 /auth/me 获取当前用户信息
    ↓
LoginForm.vue emit('success') → index.vue navigateTo('/dashboard')
```

## Token 存储策略

- 运行时用 `useState('auth-token')` 持有 token，避免跨组件传递。
- 持久化到 `localStorage`：
  - `irisimg_token`：JWT 字符串。
  - `irisimg_expires_at`：token 过期时间戳（Unix 秒）。
- 项目为 SPA 模式（`nuxt.config.ts` 中 `ssr: false`），无服务端渲染，`localStorage` 读写均在浏览器执行，不存在 hydration mismatch 问题。

## 请求鉴权

`useApi.ts` 创建的 `$fetch` 实例会：

1. 从 `useRuntimeConfig().public.apiBase` 读取后端基址。
2. 在 `onRequest` 中读取 `useState('auth-token')`，存在则附加请求头：
   ```
   Authorization: Bearer <token>
   ```
3. 在 `onResponse` 中解析统一响应体 `{ code, message, data }`：
   - `code === 0`：将响应替换为 `data`。
   - `code !== 0`：抛出 `ApiError(code, message)`。
4. 在 `onResponseError` 中捕获 HTTP 401，调用 `useAuth().logout()` 并跳转 `/`。

## 401 / 过期处理

- 登录失败：后端返回 HTTP 401 + `code=40100`，`login()` 抛出错误，`LoginForm.vue` 展示后端 message。
- 受保护接口返回 401：`useApi.ts` 自动清 token 并跳转登录页。
- token 过期：`isAuthenticated` 计算属性会返回 `false`；后台页面通过 `middleware/auth.ts` 守卫，进入时未登录则跳转 `/`。

## 安全说明

- 当前使用 `localStorage` 存储 token，适合单用户管理后台的开发与内网部署。
- 后续若暴露到公网，建议改为 `httpOnly` Cookie + 后端会话，或采用更严格的 XSS/CSRF 防护。
- 后端暂无 refresh token，token 过期后只能让用户重新登录。

## 修改建议

- 新增受保护页面时，在页面内 `definePageMeta({ middleware: 'auth' })` 即可，无需重复写 `onMounted` 校验。
- 鉴权守卫已统一收敛到 `middleware/auth.ts`（客户端执行）。
- 修改 `localStorage` key 时，务必同步更新 `useAuth.ts` 与 `auth.client.ts`。
