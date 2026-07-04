import type { FetchOptions } from 'ofetch'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

export class ApiError extends Error {
  constructor(
    public code: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export function useApi() {
  const { apiBase } = useRuntimeConfig().public

  const api = $fetch.create({
    baseURL: apiBase as string,
    onRequest({ options }) {
      const token = useState<string | null>('auth-token').value
      if (token) {
        // ofetch 在 onRequest 阶段已把 headers 规范化为 Headers 实例，
        // 这里直接用 set() 追加鉴权头，避免覆盖已有头。
        options.headers.set('Authorization', `Bearer ${token}`)
      }
    },
    onResponse({ response }) {
      const body = response._data as ApiResponse<unknown>
      if (!body || typeof body.code !== 'number') {
        return
      }
      if (body.code !== 0) {
        throw new ApiError(body.code, body.message || '请求失败')
      }
      response._data = body.data
    },
    onResponseError({ response }) {
      if (response.status === 401) {
        const { logout } = useAuth()
        logout()
        navigateTo('/')
      }
    },
  })

  async function get<T>(url: string, opts?: FetchOptions) {
    // method 放在 spread 之后，保证推断为字面量 'GET'，避免被 opts.method 放宽成 string。
    return api<T>(url, { ...opts, method: 'GET' })
  }

  async function post<T>(url: string, body?: FetchOptions['body'], opts?: FetchOptions) {
    return api<T>(url, { ...opts, method: 'POST', body })
  }

  return { api, get, post }
}
