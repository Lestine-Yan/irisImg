# `frontend/app/app.vue`

应用根组件，负责挂载全局 announcer 与页面路由出口。

## 职责

- 渲染 `NuxtRouteAnnouncer`，提供路由切换时的可访问性提示。
- 渲染 `NuxtPage`，作为文件路由的页面出口。

## 与其它文件的关系

- `app/pages/index.vue` 与 `app/pages/dashboard.vue` 通过 `NuxtPage` 在此渲染。
- 当前未使用 `NuxtLayout`，登录页通过 `definePageMeta({ layout: false })` 自行控制全屏布局。

## 修改建议

- 若后续需要统一布局（如顶部导航、侧边栏），可引入 `NuxtLayout` 并创建 `app/layouts/default.vue`。
- 登录页保持 `layout: false`，避免默认布局破坏全屏背景设计。
