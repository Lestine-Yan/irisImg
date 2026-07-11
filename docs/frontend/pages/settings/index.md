# `frontend/app/pages/settings/index.vue`

系统配置主页，对应路由 `/settings`，后台「系统配置」导航目标。只读展示当前运行配置（服务 / 数据库 / APIKey / 存储），不提供在线修改。

## 职责

- 顶部标题 + 副标题（「查看当前运行配置（只读）」）。
- 固定 amber 提示横幅：「生产中无法变更配置，请修改 config 文件并重启以应用。」
- 数据态下，当 `config.apikey.https_only === false` 时额外显示 rose 警告横幅，提示密钥相关敏感接口未强制 HTTPS，并给出 `apikey.https_only: true` 的配置建议（建议值包在 `<code>` 内，`font-mono` + `px-1.5`，与 [`ApiKeyTable`](../../components/apikeys/ApiKeyTable.md) 的 `code` 风格一致）。
- 四个 [`SettingsConfigSection`](../../components/settings/ConfigSection.md) 分组（`grid` 响应式 1/2 列）：
  - **服务**：监听地址 `host:port`。
  - **数据库**：驱动 `driver`、位置 `path`。
  - **APIKey**：默认限速 `rate_limit_per_minute 次/分钟`、HTTPS 校验徽章（`https_only` 为真显 emerald「已启用」，否则显 rose「未启用」）。
  - **存储**：根目录 `root_dir`、公访问基址 `public_base_url`（未设置时回退「未设置（使用相对路径 /imgs/）」）、上传上限 `max_upload_size_mb MiB`、允许类型 `allowed_mime_types`（`iris-violet` 徽章集合，空时显「无」）。
- 组件按目录前缀自动导入：`components/settings/ConfigSection.vue` -> `<SettingsConfigSection />`、`ConfigItem.vue` -> `<SettingsConfigItem />`（文件名未以目录名 `settings` 开头，前缀不去重）。

## 鉴权逻辑

- `definePageMeta({ middleware: 'auth' })`：未登录跳转 `/`。
- 配置拉取走后台 JWT 通道（`GET /system/config`），由 [`useApi`](../../composables/useApi.md) 自动附带 `Authorization` 头。

## 状态管理

| 类别 | 变量 | 类型 | 说明 |
| --- | --- | --- | --- |
| 数据 | `config` | `ref<SystemConfig \| null>` | 当前配置快照，非空时进入数据态渲染四分组。 |
| 加载 | `loading` | `ref<boolean>` | 拉取态，控制加载占位的显隐。初值 `ref(true)`，首帧即展示加载占位，消除数据返回前的空白闪烁。 |
| 错误 | `error` | `ref<string \| null>` | 错误文案，非空时进入错误态并展示重试按钮。 |

## 关键函数

| 函数 | 作用 |
| --- | --- |
| `fetchConfig()` | 调 [`useSystemConfig`](../../composables/useSystemConfig.md).`load()`，成功写入 `config`；失败写 `error`（`e.message` 或「加载配置失败」）并清空 `config`；`finally` 复位 `loading`。 |

## 数据流

- `onMounted(fetchConfig)`：首屏拉取配置，三态互斥渲染：`loading`（虚框 `border-dashed border-gray-200` + 「加载中…」）> `error`（`border-dashed border-rose-200 bg-rose-50` 卡片 + 文案 + 灰色「重试」按钮 `border-gray-200 text-gray-600 hover:bg-gray-50`，点击重跑 `fetchConfig`，样式与 [`ApiKeyTable`](../../components/apikeys/ApiKeyTable.md) 对齐）> `config`（提示横幅 + 可选 HTTPS 警告 + 四分组）。
- 页面为纯只读：无任何写回 / 热更新入口，配置变更需修改 `config` 文件并重启后端。

## 与其它文件的关系

```
pages/settings/index.vue
  ├── composables/useSystemConfig.ts           -> load（GET /system/config，JWT 通道）
  ├── components/settings/ConfigSection.vue    （<SettingsConfigSection />，分组卡片）
  └── components/settings/ConfigItem.vue       （<SettingsConfigItem />，键值行）
layouts/default.vue（承载侧边栏与用户区）
```

## 修改建议

- 新增配置段时，扩展 [`useSystemConfig`](../../composables/useSystemConfig.md) 的 `SystemConfig` 类型与后端 `model.SystemConfigResponse`，并在本页追加对应 `SettingsConfigSection` / `SettingsConfigItem`。
- 若后续开放在线修改，需另起表单页与写接口，本页保持只读快照职责不变。
