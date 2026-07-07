# `frontend/app/components/apikeys/ApiKeyTable.vue`

APIkey 列表表格，Nuxt 自动导入标签为 `<ApikeysApiKeyTable />`。

## 职责

- 四态：加载中 / 错误（含重试）/ 空态 / 数据表格。
- 表格列（从左到右）：**操作** ｜ 名称（含读写/只读、已吊销徽章）｜ 明文前缀 ｜ 创建时间 ｜ 最近使用时间。
- 操作列为 SVG 图标按钮（heroicons outline 风格）：重命名 / 重置明文 / 吊销或删除，`title` 属性提供 hover 说明。
- 已吊销的行降透明度，但仍可操作。

## Props / Emits

- props：`keys: APIKeyInfo[]`、`loading: boolean`、`error: string | null`。
- emits：`rename(key)`、`reset(key)`、`revokeDelete(key)`、`retry()`。

## 实现要点

- 时间格式化复用 [`useImages`](../../composables/useImages.md) 的 `formatDate`（自动导入）；`last_used_at` 为空显示「从未」。
- 前缀展示为 `<code>{prefix}…</code>`。

## 与其它文件的关系

- 父组件：[`pages/apikeys/index.vue`](../../pages/apikeys/index.md)。
- 类型：[`useApiKeys`](../../composables/useApiKeys.md)。
