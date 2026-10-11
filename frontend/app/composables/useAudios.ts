import type { FetchOptions } from 'ofetch'
import { resolveImageUrl } from '~/composables/useImages'

/** 单个音频元信息，对应后端 model.Audio。 */
export interface AudioItem {
  id: number
  filename: string
  stored_path: string
  url: string
  size: number
  mime_type: string
  hash: string
  format: string
  // ---- 标签元数据（缺失为空串/0） ----
  title: string
  artist: string
  album: string
  album_artist: string
  composer: string
  genre: string
  year: number
  track_number: number
  disc_number: number
  // ---- 技术参数（未知为 0） ----
  duration_seconds: number
  bitrate_kbps: number
  sample_rate: number
  channels: number
  bit_depth: number
  // ---- 双层封面：生效封面 = cover_url || embedded_cover_url ----
  embedded_cover_url: string
  cover_url?: string | null
  created_at: string
  key_id?: number | null
}

/** 音频列表分页响应，对应 GET /admin/audios 的 data。 */
export interface AudioListResponse {
  items: AudioItem[]
  total: number
  page: number
  page_size: number
}

export interface ListAudiosParams {
  /** 指定密钥 ID 时只返回该密钥添加的音频；不传则返回全部。 */
  keyId?: number
  /** 时间排序方向，默认 asc（升序）。 */
  order?: 'asc' | 'desc'
  /** 页码，从 1 开始，默认 1。 */
  page?: number
  /** 每页条数，默认 24。 */
  pageSize?: number
}

/** 批量删除音频请求体（账号密码二次确认 + 待删除 ID 列表）。 */
export interface BatchDeleteAudiosRequest {
  username: string
  password: string
  ids: number[]
}

/** 批量删除音频响应：deleted 为实际删除条数，ids 为实际被删除的音频 ID（不存在的 ID 静默跳过）。 */
export interface BatchDeleteAudiosResult {
  deleted: number
  ids: number[]
}

/**
 * useAudios 封装音频中心用到的列表请求、后台直传上传、批量删除与改封面。
 *
 * - 列表走后台 JWT 通道 GET /admin/audios（由 useApi 自动附带 Authorization 头）。
 * - 上传走后台 JWT 通道 POST /admin/audios，与对外 /audios（API Key 鉴权）解耦；
 *   上传的音频不关联密钥（key_id 留空，即 admin 直传）。后端按内容魔数校验
 *   mp3/m4a/flac，扩展名/Content-Type 伪造无效。
 * - 批量删除走后台 JWT 通道 DELETE /admin/audios，需账号密码二次确认。
 * - 改封面走后台 JWT 通道 PATCH /admin/audios/:id，cover_url 传空串表示清除
 *   覆盖（恢复展示内嵌封面），返回更新后的完整 AudioItem。
 */
export function useAudios() {
  const { get, post, api } = useApi()

  async function list(params: ListAudiosParams = {}): Promise<AudioListResponse> {
    const query: Record<string, string> = {
      order: params.order ?? 'asc',
      page: String(params.page ?? 1),
      page_size: String(params.pageSize ?? 24),
    }
    if (params.keyId != null) {
      query.key_id = String(params.keyId)
    }
    return get<AudioListResponse>('/admin/audios', { query } as FetchOptions)
  }

  /**
   * upload 经后台 JWT 通道上传一个音频（admin 直传，不关联密钥）。
   *
   * - 走 POST /admin/audios，FormData 字段名固定 "file"（与后端 uploadFormField 一致）。
   * - ofetch 对 FormData body 自动设置 multipart/form-data 边界，切勿手动设 Content-Type。
   * - 返回值即新增的 AudioItem（key_id 为 null，含解析出的元数据与封面）。
   */
  async function upload(file: File): Promise<AudioItem> {
    const fd = new FormData()
    fd.append('file', file)
    return post<AudioItem>('/admin/audios', fd)
  }

  /**
   * batchRemove 经后台 JWT 通道批量删除音频（物理文件 + 记录同删，不可恢复；
   * 内嵌封面为内容寻址共享文件，刻意不随删除清理）。
   *
   * - 走 DELETE /admin/audios，body 携带账号密码（后端二次确认）与待删除 ID 列表
   *   （后端限制非空、每项 > 0、上限 100）。
   * - 二次确认失败后端返回 403（而非 401），不会触发 useApi 的全局登出，
   *   调用方可在弹窗内联展示错误并让用户重试。
   */
  async function batchRemove(params: BatchDeleteAudiosRequest): Promise<BatchDeleteAudiosResult> {
    return api<BatchDeleteAudiosResult>('/admin/audios', { method: 'DELETE', body: params })
  }

  /**
   * updateCover 修改单个音频的封面 URL。
   * coverUrl 传空串表示清除用户覆盖、恢复展示内嵌封面；返回更新后的 AudioItem，
   * 调用方据此就地替换列表项（无需整页刷新）。
   */
  async function updateCover(id: number, coverUrl: string): Promise<AudioItem> {
    return api<AudioItem>(`/admin/audios/${id}`, { method: 'PATCH', body: { cover_url: coverUrl } })
  }

  return { list, upload, batchRemove, updateCover }
}

/**
 * effectiveCover 计算生效封面：用户覆盖 cover_url 优先，其次内嵌封面
 * embedded_cover_url（上传时从音频标签抽取），都没有返回空串（UI 显示占位图）。
 */
export function effectiveCover(item: AudioItem): string {
  return item.cover_url || item.embedded_cover_url || ''
}

/**
 * resolveAudioUrl 把后端返回的音频 URL 解析成浏览器可加载的完整地址。
 * 语义与 resolveImageUrl 相同（相对路径 /audios/... 拼后端 origin，绝对地址原样返回）。
 */
export function resolveAudioUrl(url: string | undefined | null): string {
  return resolveImageUrl(url)
}

/**
 * formatDuration 把秒数格式化为「m:ss」（≥1 小时为「h:mm:ss」）。
 * 0 / 空 / 非法值返回 "—"（后端未知时 duration_seconds 为 0）。
 */
export function formatDuration(seconds: number | undefined | null): string {
  if (!seconds || seconds <= 0 || !Number.isFinite(seconds)) return '—'
  const total = Math.round(seconds)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  const pad = (x: number) => String(x).padStart(2, '0')
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`
}
