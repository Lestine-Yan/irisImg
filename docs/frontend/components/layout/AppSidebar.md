# `frontend/app/components/layout/AppSidebar.vue`

后台左侧导航栏组件，Nuxt 自动导入名为 `<LayoutAppSidebar>`，由 `layouts/default.vue` 使用。

## 职责

- 顶部展示 iris 花形标记与 `irisImg / private cloud` 字标；窄屏只保留标记。
- 中部按三组渲染导航：Overview / Library / System，数据驱动。
- 底部嵌入 `<LayoutUserBadge>` 与 `self-hosted` 运行状态。
- 当前路由对应项使用白色玻璃底 + 左侧紫罗兰细条 + 图标变色。

## 导航分组

`navGroups` 数组，每组含 `label` 与 `items`；每项含 `label` / `to` / `icon`（heroicons outline 风格 SVG path）。

| 分组 | 标签 | 路由 |
|---|---|---|
| Overview | 仪表盘 | `/dashboard` |
| Library | 图片 | `/content/img` |
| Library | 音频 | `/content/audio` |
| System | 日志中心 | `/logs` |
| System | APIkey 管理 | `/apikeys` |
| System | 系统配置 | `/settings` |

> 原「内容中心 `/content`」已拆分为图片 / 音频两个平级入口；旧 `/content` 路径经 [`pages/content/index.vue`](../../pages/content/index.md) 重定向到 `/content/img`，重定向瞬间导航无高亮属预期。

## 视觉与响应式

- 根节点：`sticky top-0 z-20 h-screen w-[76px] xl:w-[264px]`，半透明 `bg-white/55` + `backdrop-blur-2xl` + 右侧 `#1C1A27/6%` 细线。
- `< xl`：76px 窄栏，只显示图标；导航项居中，`title` 提供 hover 提示，底部用户区只保留头像与退出按钮。
- `>= xl`：264px 完整栏，显示品牌文字、分组标签、导航文案、用户信息与 `self-hosted` 状态。
- 导航项基础：`rounded-xl px-3 py-2.5 text-[13px]`；active 为 `bg-white/80 text-[#1C1A27]` + 柔和阴影 + 左侧 2px 紫条；非 active 为 `text-[#6B6775]`，hover 转 `bg-white/50`。
- active 判定：`route.path === to` 或 `route.path.startsWith(to + '/')`；同时写入 `aria-current="page"`。

## 与其它文件的关系

```
AppSidebar.vue
  ├── useRoute()            # active 态判定
  └── UserBadge.vue         # 底部用户区
layouts/default.vue         # 外层背景与内容区
```

## 修改建议

- 新增路由时往对应分组的 `items` 追加一项即可，无需改模板。
- 图标可后续抽取 `AppIcon.vue` 或引入 `@nuxt/icon`。
- 窄屏目前固定 76px，不做抽屉；若需要移动端全屏导航，可在此组件外加汉堡按钮与抽屉。
