export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  token_type: string
  expires_at: number
}

export interface UserProfile {
  username: string
}

const TOKEN_KEY = 'irisimg_token'
const EXPIRES_KEY = 'irisimg_expires_at'

export function useAuth() {
  const token = useState<string | null>('auth-token', () => null)
  const expiresAt = useState<number | null>('auth-expires', () => null)
  const user = useState<UserProfile | null>('auth-user', () => null)
  const { post, get } = useApi()

  const isAuthenticated = computed(() => {
    if (!token.value) return false
    if (expiresAt.value && Date.now() / 1000 > expiresAt.value) return false
    return true
  })

  function setAuth(data: LoginResponse) {
    token.value = data.token
    expiresAt.value = data.expires_at
    if (import.meta.client) {
      localStorage.setItem(TOKEN_KEY, data.token)
      localStorage.setItem(EXPIRES_KEY, String(data.expires_at))
    }
  }

  function clearAuth() {
    token.value = null
    expiresAt.value = null
    user.value = null
    if (import.meta.client) {
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(EXPIRES_KEY)
    }
  }

  async function login(username: string, password: string) {
    const data = await post<LoginResponse>('/auth/login', { username, password } as LoginRequest)
    setAuth(data)
    await fetchMe()
  }

  async function fetchMe() {
    if (!token.value) return
    try {
      const data = await get<UserProfile>('/auth/me')
      user.value = data
    } catch {
      clearAuth()
    }
  }

  function logout() {
    clearAuth()
  }

  function initAuth() {
    if (!import.meta.client) return
    const savedToken = localStorage.getItem(TOKEN_KEY)
    const savedExpires = localStorage.getItem(EXPIRES_KEY)
    if (savedToken && savedExpires) {
      const exp = Number(savedExpires)
      if (Date.now() / 1000 < exp) {
        token.value = savedToken
        expiresAt.value = exp
        fetchMe()
      } else {
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(EXPIRES_KEY)
      }
    }
  }

  return {
    token,
    expiresAt,
    user,
    isAuthenticated,
    login,
    logout,
    fetchMe,
    initAuth,
  }
}
