// https://nuxt.com/docs/api/configuration-nuxt-config
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
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1',
    },
  },
})
