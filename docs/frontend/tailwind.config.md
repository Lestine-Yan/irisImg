# `frontend/tailwind.config.ts`

Tailwind CSS 配置文件。

## 职责

- 扩展 Tailwind 默认主题，定义项目品牌色与字体。

## 扩展项

- `colors`（蓝紫鸢尾花色系，全局主色系）：
  - `iris-dark`：`#6D4FD8` — 深紫罗兰，主按钮 / 文字 / 深底；亦是字标 `iris` 的色值。
  - `iris-violet`：`#A48CE6` — 浅紫罗兰，字标 `Img` 的色值 / 辅助强调。
  - `iris-sky`：`#86C9EC` — 浅天蓝，渐变底色。
  - `iris-gold`：`#F4C430` — 金黄，鸢尾斑纹强调色。
  - `iris-cream`：`#FAFBE6` — 暖米白，背景留白底色。
- `fontFamily`：优先使用 `Inter`，回退到系统字体。

## 与其它文件的关系

- 被 `@nuxtjs/tailwindcss` 模块读取，决定生成的工具类。
- 品牌色在 `frontend/app/components/login/GeometricBackground.vue`（外层 `bg-iris-cream`，SVG 内联色值与之一致）、`LoginHero.vue`（字标 `iris` 与 `Img` 均为 `iris-violet`）与 `LoginForm.vue`（主按钮 `iris-dark` 底 + 白字、hover 底色转 `iris-violet`；输入框聚焦态 `iris-dark`）中体现。

## 配置加载与排错

- 本文件**未设置 `content`**：内容扫描路径由 `@nuxtjs/tailwindcss` 模块自动注入（`app/components/**`、`app/pages/**`、`app/layouts/**`、`app/composables/**` 等，见 `.nuxt/tailwind/postcss.mjs`）。不要写 `content: []`，否则在某些纯 CLI 场景会误判为「不扫描任何文件」。
- **改完 `tailwind.config.ts` 必须重启 `pnpm dev`**：模块对 `.ts` 配置的 HMR 不可靠（`builder:watch` 仅在开启 `exposeConfig` 时触发），运行中的 dev server 不会自动应用新的色值 / token，会继续下发旧 CSS，表现为「全局颜色没被应用」。若遇此现象，先停掉旧 dev server（留意 Nuxt dev 锁提示的 PID）再重新 `pnpm dev`。
- `frontend/app/assets/css/tailwind.css` 中的 `@tailwind base/components/utilities` 是 Tailwind v3 的标准入口；VS Code 自带 CSS 语言服务会报 `Unknown at rule @tailwind`，属误报，不影响构建。已通过 `frontend/.vscode/settings.json` 的 `css.lint.unknownAtRules: "ignore"` 静默。

## 修改建议

- 新增颜色/间距/阴影变量时，优先在此集中定义，避免在组件中写死色值。
- `iris-violet`（`#A48CE6`）偏浅，字标与按钮 hover 刻意采用以追求观感（对比度约 2.7:1，偏低但可接受）；需要高对比的文字或按钮主态仍用 `iris-dark`（`#6D4FD8`）。
- `GeometricBackground.vue` 中的渐变 / 留白色值当前直接写死并与本配置保持一致；若后续需要统一管理，可考虑改为 CSS 变量引用。
- 若后续引入 `tailwindcss-animate` 等插件，在此 `plugins` 数组注册。
