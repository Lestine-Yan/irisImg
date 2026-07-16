# `frontend/app/components/logs/LogsHistogram.vue`

按日计数直方图（纯 SVG、零第三方绘图依赖；默认渲染近 14 天日志量，亦可复用于仪表盘近 N 天新增图片趋势），Nuxt 自动导入标签为 `<LogsHistogram />`（`components/logs/LogsHistogram.vue` 文件名以目录名 `logs` 开头，Nuxt 去重前缀，注册名 `LogsHistogram` 而非 `LogsLogsHistogram`）。

## 职责

- 四态：加载中 / 错误态（`error` 非空，rose 虚框 + 错误文案 + 重试按钮 emit `retry`）/ 空态（`total === 0`，显示 `emptyText`，默认「近 14 天暂无日志」，仪表盘图片趋势传「近 30 天暂无新增图片」）/ 直方图 + 趋势线。渲染优先级：`loading` > `error` > `total === 0` > 直方图。
- SVG 画布 `viewBox="0 0 700 240"`，内边距常量 `PAD_L=40 / PAD_R=16 / PAD_T=16 / PAD_B=36`。
- 14 根柱（由 `buckets.length` 决定）：默认 `iris-violet`（`#A48CE6`），hover 切换 `iris-dark`（`#6D4FD8`）；每根柱带 `<title>` 浮窗（`{date}：{count} 条`），并在柱顶显示当前数值。
- 7 日移动平均趋势线：`iris-gold`（`#F4C430`）`polyline` + 圆点；窗口不足 7 时取已有天均值，使趋势线从首日即横跨全宽。
- 4 等分网格线 + Y 轴刻度（0 / 1/4·max / 1/2·max / 3/4·max / max）。
- X 轴日期标签按自适应步长显示（`labelStep = ceil(n/8)`：14 天 step=2 与历史一致、30 天 step=4，末根恒显示），格式 `MM-DD`（`date.slice(5)`）。
- 底部图例：紫色方块 `legendText`（默认「日志量」）+ 金色短线「7 日均值」。

## Props / Emits

- props：`buckets: HistogramBucket[]`、`total: number`、`loading: boolean`、`error: string | null`、`emptyText?: string`（空态文案，默认「近 14 天暂无日志」）、`titleText?: string`（`aria-label` 文案，默认「近 14 天日志量直方图」）、`legendText?: string`（柱状图图例文案，默认「日志量」）。`emptyText` / `titleText` / `legendText` 使组件可复用于图片趋势等非日志场景，默认值保持日志中心原行为不变。
- emits：`retry: []`（仅在错误态下点击「重试」按钮时触发，父组件据此重新拉取直方图数据）。

## 实现要点

- 零依赖：仅用原生 SVG 元素（`rect` / `polyline` / `circle` / `line` / `text`）与 Vue 模板渲染，未引入 ECharts/D3 等。
- `max` 取 `Math.max(1, ...counts)`，至少为 1 以避免除零；无数据时走空态分支不渲染柱。
- 柱布局按等分 `slot = plotW / n`，柱宽 `slot * 0.6`，圆角 `rx="3"`。
- 趋势点复用 `bars[i].cx` 作为横坐标，纵坐标按 `v / max * plotH` 反向映射到画布。
- 配色沿用 iris 色系：柱 `iris-violet` / `iris-dark`、趋势线 `iris-gold`、图例方块 `bg-iris-violet`、短线 `bg-iris-gold`；网格线用浅灰 `#F1F5F9`、文字用 `fill-gray-400`。

## 与其它文件的关系

- 父组件：[`pages/logs/index.vue`](../../pages/logs/index.md)（近 14 天日志量）、[`pages/dashboard.vue`](../../pages/dashboard.md)（近 30 天新增图片趋势，传 `empty-text` / `title-text` 适配）。
- 类型来源：[`composables/useLogs`](../../composables/useLogs.md)（`HistogramBucket`，仪表盘经 [`useDashboard`](../../composables/useDashboard.md) 复用同结构）。
