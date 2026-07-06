# `frontend/app/components/layout/AppSidebar.vue`

后台左侧导航栏组件，Nuxt 自动导入名为 `<LayoutAppSidebar>`，由 `layouts/default.vue` 使用。

## 职责

- 顶部展示 `irisImg` 字标（`iris-violet` 色，与 `LoginHero` 字标一致）。
- 中部渲染导航项列表，数据驱动，`v-for` 输出 `NuxtLink`。
- 底部嵌入 `<LayoutUserBadge>` 展示当前用户。

## 导航项

内置 `navItems` 数组，每项含 `label` / `to` / `icon`（heroicons outline 风格 SVG path 字符串）：

| 标签 | 路由 | 图标 |
|---|---|---|
| 仪表盘 | `/dashboard` | grid |
| 内容中心 | `/content` | photo |
| 日志中心 | `/logs` | document |
| APIkey 管理 | `/apikeys` | key |
| 系统配置 | `/settings` | cog |

## 视觉

- 白底 + 右侧细边框（`bg-white border-r border-gray-200`），宽度 `w-64`，`sticky top-0 h-screen`。
- 导航项默认 `text-gray-600`；hover `bg-gray-100 text-gray-900`；active `bg-iris-violet/15 text-iris-dark`。
- active 判定：`route.path === to` 或 `route.path.startsWith(to + '/')`。
- 当前在所有断点固定显示（不随视口隐藏），暂未做移动端汉堡菜单适配，小屏下会与内容区挤压。

## 与其它文件的关系

```
AppSidebar.vue
  ├── useRoute()            # active 态判定
  └── UserBadge.vue         # 底部用户区
```

## 修改建议

- 新增路由时往 `navItems` 追加一项即可，无需改模板。
- 移动端后续可加汉堡按钮 + 抽屉式展开。
- 若图标增多，可提取 `AppIcon.vue` 或引入 `@nuxt/icon`。
