import type { FetchOptions } from 'ofetch'
import type { HistogramBucket } from '~/composables/useLogs'

/** 仪表盘聚合统计响应，对应 GET /admin/dashboard 的 data（后端 model.DashboardOverview）。 */
export interface DashboardStats {
  /** 图片总量。 */
  images_total: number
  /** 已登记图片总大小（字节）。 */
  storage_bytes: number
  /** 密钥总数（含已吊销；已删除为物理删除，不在统计内）。 */
  apikeys_total: number
  /** 未吊销的有效密钥数。 */
  apikeys_active: number
  /** 已吊销密钥数。 */
  apikeys_revoked: number
  /** 日志总量。 */
  logs_total: number
  /** 近 N 天每日新增图片（结构同日志直方图 buckets，可直接喂给 LogsHistogram）。 */
  recent_upload_trend: HistogramBucket[]
  /** 近 N 天新增图片合计。 */
  recent_upload_total: number
  /** 趋势窗口天数。 */
  days: number
}

/**
 * useDashboard 封装仪表盘首页用到的统计接口。
 *
 * 走后台 JWT 通道（由 useApi 自动附带 Authorization 头、自动解包 data）。
 * overview() 一次性返回图片总量 / 存储占用 / APIkey 计数 / 日志总量 / 近 N 天上传趋势。
 */
export function useDashboard() {
  const { get } = useApi()

  /** 拉取仪表盘聚合统计；days 控制趋势窗口，默认 30（后端上限 90）。 */
  async function overview(days = 30): Promise<DashboardStats> {
    return get<DashboardStats>('/admin/dashboard', {
      query: { days: String(days) },
    } as FetchOptions)
  }

  return { overview }
}
