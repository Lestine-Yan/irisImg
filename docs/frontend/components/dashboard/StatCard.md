# `frontend/app/components/dashboard/StatCard.vue`

仪表盘统计卡片：左侧标题 + 大数字 + 可选副标题，右侧 iris 图标徽章。Nuxt 自动导入标签为 `<DashboardStatCard />`（`components/dashboard/StatCard.vue` 目录名 `dashboard` 不与文件名 `StatCard` 重叠，故注册名 `DashboardStatCard`，不去重前缀）。

## 职责

- 纯展示组件，把一个统计指标渲染为卡片：标题（`label`）、主数值（`value`，字符串或数字，由父组件格式化后传入）、可选副标题（`hint`）、可选图标徽章（`icon`，heroicons outline 的 SVG path d）。
- 容器样式沿用 `ConfigSection` 的 `rounded-2xl border border-gray-200 bg-white p-6` 范式。
- 图标徽章用 `bg-iris-violet/10` 底 + `text-iris-dark` 字色，与侧边栏激活态呼应。

## Props

- `label: string`：指标标题（如「图片总量」）。
- `value: string | number`：主数值；存储大小等需格式化的指标由父组件用 `formatBytes` 等转成字符串后传入。
- `icon?: string`：heroicons outline path d；不传则不渲染图标徽章。
- `hint?: string`：副标题（如「有效 2 · 已吊销 1」）；不传则不渲染。

## 实现要点

- 无 emits、无内部状态，纯 props 驱动。
- 图标 SVG 沿用全站写法：`fill="none" stroke="currentColor" viewBox="0 0 24 24"`，`stroke-width="1.5"`，path d 由 `icon` 传入。
- 配色仅用 iris token + gray 默认色 + 透明度修饰符，不引入自定义颜色。

## 与其它文件的关系

- 父组件：[`pages/dashboard.vue`](../../pages/dashboard.md)（渲染 4 张统计卡）。
- 样式参考：[`components/settings/ConfigSection.vue`](../settings/ConfigSection.md)（白底圆角卡范式）。
