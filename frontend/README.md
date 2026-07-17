# irisImg frontend

基于 **Nuxt 4 + Vue 3** 的图床前端，与 [`backend/`](../backend) 中的 Gin 后端配合使用。

## 技术栈

- [Nuxt 4](https://nuxt.com/) — Vue 全栈框架
- [Vue 3](https://vuejs.org/) — 组合式 API
- [Vue Router 5](https://router.vuejs.org/)
- [Tailwind CSS](https://tailwindcss.com/) — 原子化 CSS
- TypeScript
- pnpm — 包管理器

## 目录结构

```
frontend/
├── app/
│   ├── app.vue              # 应用根组件
│   ├── assets/css/          # Tailwind CSS 入口
│   ├── pages/               # 文件路由
│   ├── components/login/    # 登录页组件
│   ├── composables/         # 自动导入的组合函数
│   └── plugins/             # 客户端插件
├── public/                  # 静态资源（favicon、robots.txt 等）
├── nuxt.config.ts           # Nuxt 配置
├── tailwind.config.ts       # Tailwind 主题扩展
├── tsconfig.json
├── package.json
└── pnpm-lock.yaml
```

> Nuxt 约定式目录（按需创建）：
> - `app/pages/` — 文件路由
> - `app/components/` — 自动注册的组件
> - `app/composables/` — 自动导入的组合函数
> - `app/layouts/` — 布局
> - `app/middleware/` — 路由中间件
> - `app/assets/` — 需要构建处理的资源（CSS、图片等）
> - `server/api/` — Nuxt 服务端接口（如不需要可不建）

## 环境要求

- Node.js ≥ 20
- pnpm ≥ 9（推荐）

## 快速开始

```bash
cd frontend
pnpm install
pnpm dev
```

默认开发服务器地址：<http://localhost:3000>

## 常用脚本

| 命令              | 说明                                |
| ----------------- | ----------------------------------- |
| `pnpm dev`        | 启动开发服务器（带 HMR）            |
| `pnpm build`      | 构建生产版本（SSR 输出到 `.output`）|
| `pnpm generate`   | 生成纯静态站点（SSG）               |
| `pnpm preview`    | 本地预览生产构建                    |
| `pnpm postinstall`| `nuxt prepare`（一般无需手动执行）  |

## 登录

访问 `/` 进入登录页。输入后端配置的用户名与密码后，前端会：

1. 调 `POST /api/v1/auth/login` 获取 JWT。
2. 将 token 与过期时间写入 `localStorage`。
3. 跳转 `/dashboard`。

登录态由 `app/composables/useAuth.ts` 维护，请求封装见 `app/composables/useApi.ts`。详细说明参见 `docs/frontend/AUTH.md`。

## 与后端联调

后端默认监听 `http://localhost:8080`，所有业务接口挂在 `/api/v1` 下。

`nuxt.config.ts` 已通过 `runtimeConfig` 暴露 API 基址（`sanitizeApiBase` 用于规避 Git Bash/MSYS2 路径转换污染，详见 [`docs/deploy.md`](../docs/deploy.md)「Git Bash 构建坑」）：

```ts
runtimeConfig: {
  public: {
    apiBase: sanitizeApiBase(process.env.NUXT_PUBLIC_API_BASE) || 'http://localhost:8080/api/v1',
  },
}
```

在组件里使用：

```ts
const { apiBase } = useRuntimeConfig().public
const { data } = await useFetch(`${apiBase}/ping`)
```

### 跨域

后端已挂载允许全部来源的 CORS 中间件（见 `backend/internal/middleware/cors.go`），开发环境可直接调用。
若想走 Nuxt 代理避免跨域，也可以在 `nuxt.config.ts` 配置 `nitro.devProxy`：

```ts
nitro: {
  devProxy: {
    '/api': { target: 'http://localhost:8080/api', changeOrigin: true },
  },
}
```

## 部署

按项目计划，最终前端会以 **静态资源** 形式由 Nginx 直接代理：

```bash
pnpm generate          # 输出到 .output/public
```

将 `.output/public/` 拷贝到 Nginx 站点目录即可，后端仍由 Gin 可执行文件单独运行。

## 后续可演进方向

- 补齐上传、图片管理、设置等页面
- 引入 UI 库（Element Plus / Naive UI / shadcn-vue 任选）
- 接入 Pinia 管理更复杂的全局状态
- 接入 ESLint + Prettier 统一代码风格
