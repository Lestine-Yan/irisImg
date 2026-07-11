/** 系统配置只读视图，对应后端 model.SystemConfigResponse。 */
export interface SystemConfig {
  server: {
    host: string
    port: number
  }
  database: {
    driver: string
    path: string
  }
  apikey: {
    rate_limit_per_minute: number
    https_only: boolean
  }
  storage: {
    root_dir: string
    public_base_url: string
    max_upload_size_mb: number
    allowed_mime_types: string[]
  }
}

/**
 * useSystemConfig 封装系统配置只读接口。
 *
 * 走后台 JWT 通道（由 useApi 自动附带 Authorization 头、自动解包 data）。
 * 接口仅返回 config 的非敏感快照，前端不做任何修改 / 热更新。
 */
export function useSystemConfig() {
  const { get } = useApi()

  async function load(): Promise<SystemConfig> {
    const cfg = await get<SystemConfig>('/system/config')
    // 防御性兜底：后端已保证 allowed_mime_types 非 null，此处再保险，
    // 避免后端契约变动时模板里 .length / v-for 抛空指针。
    if (cfg.storage.allowed_mime_types == null) {
      cfg.storage.allowed_mime_types = []
    }
    return cfg
  }

  return { load }
}
