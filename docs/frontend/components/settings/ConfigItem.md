# `frontend/app/components/settings/ConfigItem.vue`

系统配置键值行，Nuxt 自动导入标签为 `<SettingsConfigItem />`（组件路径 `components/settings/ConfigItem.vue`，目录前缀 `Settings` + 文件名 `ConfigItem`，文件名未以目录名 `settings` 开头故不去重前缀）。

## 职责

- 渲染 `dl` 内的一行：左侧 `dt` 为字段标签（灰色小字），右侧 `dd` 为字段值（右对齐深色中字），二者 `flex items-start justify-between gap-4`，行高由 `py-3` 提供。
- 值通过默认插槽注入，调用方可传入纯文本或任意节点（徽章、标签集合等）。

## Props / Emits

- props：`label: string`（字段标签，如「监听地址」「驱动」「HTTPS 校验」）。
- emits：无。

## 实现要点

- 无业务逻辑，仅 `defineProps<{ label: string }>()` + 模板。
- 值列 `text-right`，配合 [`ConfigSection`](./ConfigSection.md) 的 `divide-y` 形成整齐的键值表；调用方负责空值回退（如存储 `public_base_url` 未设置时由页面显示「未设置（使用相对路径 /imgs/）」）。

## 与其它文件的关系

- 父组件：[`pages/settings/index.vue`](../../pages/settings/index.md)，通常作为 [`ConfigSection`](./ConfigSection.md) 默认插槽的子项。
