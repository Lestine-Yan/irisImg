// 客户端路由守卫：未登录用户访问后台页面时跳转到登录页。
//
// 项目为 SPA 模式（nuxt.config.ts 中 ssr: false），无服务端渲染，
// middleware 仅在浏览器执行；token 由 plugins/auth.client.ts 在启动时从
// localStorage 恢复，运行到此处时登录态已就绪。
export default defineNuxtRouteMiddleware(() => {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated.value) {
    return navigateTo('/')
  }
})
