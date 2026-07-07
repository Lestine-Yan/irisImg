import type { FetchOptions } from 'ofetch'

/** 密钥权限范围。 */
export type APIKeyScope = 'readonly' | 'readwrite'

/** 密钥列表项，对应后端 model.APIKeyInfo（不含明文与哈希）。 */
export interface APIKeyInfo {
  id: number
  name: string
  prefix: string
  scope: APIKeyScope
  rate_limit: number
  revoked: boolean
  last_used_at: string | null
  created_at: string
}

/** 创建密钥响应，含一次性明文 key。 */
export interface CreateAPIKeyResponse {
  id: number
  name: string
  prefix: string
  scope: APIKeyScope
  key: string
  rate_limit: number
  created_at: string
}

/** 重置密钥响应，含一次性新明文 key。 */
export interface ResetAPIKeyResponse {
  id: number
  name: string
  prefix: string
  key: string
  revoked: boolean
  created_at: string
}

/** 吊销 / 删除密钥这类敏感操作的请求体（账号密码二次确认）。 */
export interface DestructiveAPIKeyRequest {
  username: string
  password: string
}

/** 删除密钥响应，附带被级联删除的图片数量。 */
export interface DeleteAPIKeyResult {
  id: number
  deleted: boolean
  images_removed: number
}

export interface CreateAPIKeyParams {
  name: string
  scope: APIKeyScope
}

/**
 * useApiKeys 封装 APIkey 管理页用到的全部接口。
 *
 * 所有接口走后台 JWT 通道（由 useApi 自动附带 Authorization 头、自动解包 data）。
 * - 创建 / 重置返回的明文 key 仅此一次，调用方需提示用户立即复制保存。
 * - 吊销 / 删除为敏感操作，需在请求体里携带账号密码做二次确认；
 *   后端密码校验失败返回 403（而非 401），避免触发 useApi 的全局登出。
 */
export function useApiKeys() {
  const { get, post, api } = useApi()

  async function list(): Promise<APIKeyInfo[]> {
    const data = await get<{ items: APIKeyInfo[] }>('/apikeys')
    return data.items ?? []
  }

  async function create(params: CreateAPIKeyParams): Promise<CreateAPIKeyResponse> {
    return post<CreateAPIKeyResponse>('/apikeys', params)
  }

  async function rename(id: number, name: string): Promise<APIKeyInfo> {
    return api<APIKeyInfo>(`/apikeys/${id}`, { method: 'PATCH', body: { name } })
  }

  async function reset(id: number): Promise<ResetAPIKeyResponse> {
    return post<ResetAPIKeyResponse>(`/apikeys/${id}/reset`)
  }

  async function revoke(id: number, creds: DestructiveAPIKeyRequest): Promise<void> {
    await post(`/apikeys/${id}/revoke`, creds)
  }

  async function purge(id: number, creds: DestructiveAPIKeyRequest): Promise<DeleteAPIKeyResult> {
    return api<DeleteAPIKeyResult>(`/apikeys/${id}`, { method: 'DELETE', body: creds })
  }

  return { list, create, rename, reset, revoke, purge }
}
