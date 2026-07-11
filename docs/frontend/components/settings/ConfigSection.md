# `frontend/app/components/settings/ConfigSection.vue`

系统配置分组卡片，Nuxt 自动导入标签为 `<SettingsConfigSection />`（组件路径 `components/settings/ConfigSection.vue`，目录前缀 `Settings` + 文件名 `ConfigSection`，文件名未以目录名 `settings` 开头故不去重前缀）。

## 职责

- 渲染一个圆角白底卡片（`rounded-2xl border bg-white p-6`），顶部为分组标题（`uppercase` 小字灰色），下方为 `dl` 列表（`divide-y divide-gray-100` 行间分隔）。
- 通过默认插槽接收若干 [`ConfigItem`](./ConfigItem.md)（`dt` / `dd` 行）作为列表内容。

## Props / Emits

- props：`title: string`（分组标题，如「服务」「数据库」「APIKey」「存储」）。
- emits：无。

## 实现要点

- 无业务逻辑，仅 `defineProps<{ title: string }>()` + 模板。
- 列表用语义化 `<dl>`，分隔线由 Tailwind `divide-y` 提供；行（`ConfigItem`）自带 `py-3`，本组件不再控制行间距。

## 与其它文件的关系

- 父组件：[`pages/settings/index.vue`](../../pages/settings/index.md)。
- 子插槽内容：[`ConfigItem`](./ConfigItem.md)。
