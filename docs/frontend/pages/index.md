# `frontend/app/pages/index.vue`

登录落地页，对应路由 `/`。

## 职责

- 组合登录页的三个核心视觉组件：
  - `LoginGeometricBackground`：全屏 SVG 几何背景。
  - `LoginHero`：左侧品牌文案与功能标签。
  - `LoginForm`：右侧白色登录卡片。
- 登录成功后跳转 `/dashboard`。

## 布局

- 使用 `definePageMeta({ layout: false })` 禁用默认布局，确保背景全屏。
- 响应式：移动端上下堆叠（上方字标、下方表单），整体居中；桌面端左右分栏。
- 桌面端左侧 Hero 列通过 `lg:self-start lg:pt-24` 顶部对齐，使 `irisImg` 字标落在背景白色留白区的左上方；右侧表单列仍垂直居中。

## 与其它文件的关系

```
index.vue
  ├── LoginGeometricBackground.vue
  ├── LoginHero.vue
  └── LoginForm.vue
        └── 登录成功 emit('success') → navigateTo('/dashboard')
```

## 修改建议

- 若后续登录页改为独立路由 `/login`，需要同步调整此处路由与 `dashboard.vue` 的跳转目标。
- 背景组件固定定位，`index.vue` 的内容容器需要设置 `relative z-10` 以覆盖背景。
