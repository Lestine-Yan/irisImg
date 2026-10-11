# `frontend/app/pages/content/index.vue`

内容中心路由拆分后的**旧链接兼容重定向桩**，对应路由 `/content`，挂载后立即 `navigateTo('/content/img', { replace: true })`，把旧书签 / 旧入口即达图片中心。

## 职责

- 仅做重定向，无任何 UI 与数据请求（模板是空 `<div />`）。
- `definePageMeta({ middleware: 'auth' })`：未登录仍先跳 `/`，登录后回到本页才继续重定向。

## 为什么是 `index.vue` 而不是 `content.vue`

**必须放在 `pages/content/index.vue`**：若建 `pages/content.vue`，Nuxt 会把它注册为 `content/img.vue` 与 `content/audio.vue` 的**嵌套父路由组件**——父组件缺少 `<NuxtPage />` 时子页面将渲染空白。`index.vue` 与 `img.vue`/`audio.vue` 是平级路由，无嵌套问题。

## 与其它文件的关系

- 重定向目标：[`pages/content/img.vue`](./img.md)（图片中心）。
- 真正的内容页：[`img.vue`](./img.md)、[`audio.vue`](./audio.md)。
