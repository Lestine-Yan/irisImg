# `frontend/app/components/logs/LogsTable.vue`

日志列表表格，Nuxt 自动导入标签为 `<LogsTable />`（文件名以目录名 `logs` 开头，Nuxt 去重前缀，注册名 `LogsTable` 而非 `LogsLogsTable`）。

## 职责

- 四态：加载中 / 错误（含重试）/ 空态 / 数据表格。
- 表格列（从左到右）：**时间** ｜ 级别徽章 ｜ 事件 ｜ 方法 · 路径 ｜ 状态 ｜ 耗时 ｜ 来源 ｜ 详情。
- 级别徽章按 `levelClass` 着色：error rose / warn amber / debug gray / info sky。
- 来源列展示优先级：`username` > `Key #{api_key_id}` > `-`。
- 详情列对 `message` 做 `truncate` 截断，hover 显示完整内容。

## Props / Emits

- props：`logs: LogItem[]`、`loading: boolean`、`error: string | null`。
- emits：`retry()`。

## 实现要点

- 时间格式化复用 [`useImages`](../../composables/useImages.md) 的 `formatDate`（显式 `import` 引入）。
- 级别 / 事件 / 方法均以 `<code>` 等宽样式呈现；`method` 为空时方法·路径列回退为 `-`。
- 状态、耗时为空时统一回退为 `-`，耗时非空时追加 `ms` 后缀。
- 类型 `LogItem`、`LogLevel` 来自 [`useLogs`](../../composables/useLogs.md)。

## 与其它文件的关系

- 父组件：[`pages/logs/index.vue`](../../pages/logs/index.md)。
- 类型：[`useLogs`](../../composables/useLogs.md)。
- 工具：[`useImages`](../../composables/useImages.md)（`formatDate`）。
