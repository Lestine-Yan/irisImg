# `frontend/nuxt.config.ts`

Nuxt 应用配置入口。

## 关键配置

- `ssr: false`：关闭服务端渲染，走 SPA 模式。后台管理系统无需 SEO / 首屏，登录态基于浏览器 `localStorage`，SSR 拿不到 token 会渲染空壳，故关闭。
- `modules: ['@nuxtjs/tailwindcss']`：接入 Tailwind，全局样式见 `assets/css/tailwind.css`，色系定义见 `tailwind.config.ts`。
- `css: ['~/assets/css/tailwind.css']`：全局引入 Tailwind 基础样式。
- `runtimeConfig.public.apiBase`：后端 API 基址，默认 `http://localhost:8080/api/v1`，可通过 `NUXT_PUBLIC_API_BASE` 覆盖。
- `devtools: { enabled: true }`：开发态启用 Nuxt DevTools。
- `compatibilityDate`：Nuxt 兼容性日期锚点。

## 部署

- `pnpm build`：产出 SPA 生产包到 `.output/`。
- `pnpm generate`：产出静态站点到 `.output/public/`，供 Nginx 部署。

## 修改建议

- 若后续需要 SEO（如公开的图片展示页），可针对单路由在 `routeRules` 内开启预渲染，但需先把该路由的鉴权改为不依赖 `localStorage`。
- 新增 `runtimeConfig` 项时同步更新此处与 `.env.example`（一旦添加）。
