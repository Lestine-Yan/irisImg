// https://nuxt.com/docs/api/configuration-nuxt-config

// Git Bash (MSYS2) 在执行原生 Windows 程序（node.exe）时，会把以 / 开头的环境变量值
// 自动转换成 Windows 绝对路径（如 /api/v1 -> D:/Runer/Git/Git/api/v1），污染构建产物，
// 导致部署后浏览器请求非法 URL 而 Failed to fetch。这里对「盘符开头」的被污染值兜底重置，
// 正常的 /api/v1、http(s)://... 等值不受影响；构建脚本另配 MSYS_NO_PATHCONV=1 双重保险。
function sanitizeApiBase(value: string | undefined): string | undefined {
  if (value && /^[A-Za-z]:[\\/]/.test(value)) {
    return '/api/v1'
  }
  return value
}

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  // 后台管理系统无需 SEO / 首屏，登录态基于浏览器 localStorage，
  // 故关闭 SSR 走 SPA 模式，避免服务端拿不到 token 而渲染空壳。
  ssr: false,
  modules: ['@nuxtjs/tailwindcss'],
  css: ['~/assets/css/tailwind.css'],
  runtimeConfig: {
    public: {
      // 构建时由 NUXT_PUBLIC_API_BASE 注入；生产同域部署取相对路径 /api/v1。
      apiBase: sanitizeApiBase(process.env.NUXT_PUBLIC_API_BASE) || 'http://localhost:8080/api/v1',
    },
  },
})
