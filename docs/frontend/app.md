# `frontend/app/app.vue`

应用根组件，负责挂载全局 announcer、布局出口与页面路由出口。

## 职责

- 渲染 `NuxtRouteAnnouncer`，提供路由切换时的可访问性提示。
- 用 `NuxtLayout` 包裹 `NuxtPage`，使 `layouts/default.vue` 等布局生效。

## 与其它文件的关系

- `NuxtPage` 渲染 `app/pages/` 下的页面。
- `NuxtLayout` 根据页面 `definePageMeta` 选择布局：
  - `pages/index.vue`（登录页）声明 `layout: false`，跳过布局，自行渲染全屏背景。
  - 其余后台页（`dashboard` / `content` / `logs` / `apikeys` / `settings`）使用默认布局 `layouts/default.vue`（侧边栏 + 内容区）。

## 注意

- Nuxt 4 中存在自定义 `app.vue` 时，layouts 不会自动注入，必须在此显式写 `<NuxtLayout>`，否则 `layouts/` 下的文件不会被渲染（控制台会出现 `Your project has layouts but <NuxtLayout /> has not been detected` 警告）。

## 修改建议

- 新增全局布局（如顶部栏）时，可创建 `layouts/<name>.vue` 并在页面 `definePageMeta({ layout: '<name>' })` 引用。
- 登录页保持 `layout: false`，避免默认布局破坏全屏背景设计。
