# `frontend/app/composables/useSystemConfig.ts`

系统配置只读接口封装 + 类型定义，供 [`pages/settings/index.vue`](../pages/settings/index.md) 使用。

## 导出类型

- `SystemConfig`：系统配置只读快照，对应后端 `model.SystemConfigResponse`，按四段组织（字段 snake_case 与后端 DTO 一致）：
  - `server`：`{ host: string; port: number }`，服务监听地址与端口。
  - `database`：`{ driver: string; path: string }`，数据库驱动与连接 / 文件路径。
  - `apikey`：`{ rate_limit_per_minute: number; https_only: boolean }`，密钥默认限速与 HTTPS 校验开关。
  - `storage`：`{ root_dir: string; public_base_url: string; max_upload_size_mb: number; allowed_mime_types: string[] }`，存储根目录、公访问基址、单文件上传上限与允许的 MIME 类型白名单。

## 导出函数

### `useSystemConfig()`

返回 `{ load }`，走后台 JWT 通道（由 [`useApi`](./useApi.md) 自动附带 `Authorization` 头、自动解包 `data`）。导出名用 `load` 而非 `fetch`，避免遮蔽全局 `fetch`，与 [`useApiKeys`](./useApiKeys.md) 的 `list`、[`useLogs`](./useLogs.md) 的 `list-histogram` 等领域命名一致。

- `load()`：调 [`useApi`](./useApi.md) 的 `get<SystemConfig>('/system/config')`，返回当前运行配置的非敏感快照。接口仅用于只读展示，前端不做任何修改 / 热更新。
- `load()` 内含防御性兜底：若返回的 `cfg.storage.allowed_mime_types` 为 `null` 则置空数组，避免后端契约变动时模板里 `.length` / `v-for` 抛空指针（后端已保证非 null，此处再保险）。

## 与其它文件的关系

- 被 [`pages/settings/index.vue`](../pages/settings/index.md) 用于首屏拉取配置并渲染只读视图。
- 底层依赖 [`useApi.ts`](./useApi.md)（`get` 自动附带 JWT 并解包 `data`）。
