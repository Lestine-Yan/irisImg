# `frontend/nuxt.config.ts`

Nuxt 应用配置入口。

## 关键配置

- `ssr: false`：关闭服务端渲染，走 SPA 模式。后台管理系统无需 SEO / 首屏，登录态基于浏览器 `localStorage`，SSR 拿不到 token 会渲染空壳，故关闭。
- `modules: ['@nuxtjs/tailwindcss']`：接入 Tailwind，全局样式见 `assets/css/tailwind.css`，色系定义见 `tailwind.config.ts`。
- `css: ['~/assets/css/tailwind.css']`：全局引入 Tailwind 基础样式。
- `runtimeConfig.public.apiBase`：后端 API 基址，默认 `http://localhost:8080/api/v1`，可通过 `NUXT_PUBLIC_API_BASE` 覆盖。取值经 `sanitizeApiBase()` 清洗：若被 Git Bash (MSYS2) 路径转换污染成盘符开头的 Windows 绝对路径（如 `D:/Runer/Git/Git/api/v1`），自动重置为 `/api/v1`，避免部署后浏览器请求非法 URL 而 `Failed to fetch`；正常的 `/api/v1`、`http(s)://...` 不受影响。
- `devtools: { enabled: true }`：开发态启用 Nuxt DevTools。
- `compatibilityDate`：Nuxt 兼容性日期锚点。

## 部署

- `pnpm build`：产出 SPA 生产包到 `.output/`。
- `pnpm generate`：产出静态站点到 `.output/public/`，供 Nginx 部署。

## 构建环境注意（Git Bash / MSYS2）

在 Windows 上用 Git Bash 构建时，MSYS2 会对传给原生 Windows 程序（`node.exe`）的环境变量值做 POSIX->Windows 路径转换：以 `/` 开头的 `NUXT_PUBLIC_API_BASE=/api/v1` 会被改写成 `D:/Runer/Git/Git/api/v1`（Git 安装根即 MSYS 根）并烤进产物，部署后浏览器请求该非法 URL 直接 `Failed to fetch`，请求不会到达后端（后端日志看不到 `/api/v1/auth/login`）。

两层防护：

1. `nuxt.config.ts` 的 `sanitizeApiBase()` 对盘符开头的值兜底重置为 `/api/v1`（应用层，跨平台稳定）。
2. `scripts/build-release.sh` 构建时加 `MSYS_NO_PATHCONV=1` 关闭转换，并对产物 grep 校验「盘符路径 + api/v1」，命中则中止。

手动构建也可在 PowerShell / cmd 里执行（不做路径转换）规避。详见 [`docs/deploy.md`](../deploy.md)「Git Bash 构建坑」。

## 修改建议

- 若后续需要 SEO（如公开的图片展示页），可针对单路由在 `routeRules` 内开启预渲染，但需先把该路由的鉴权改为不依赖 `localStorage`。
- 新增 `runtimeConfig` 项时同步更新此处与 `.env.example`（一旦添加）。
