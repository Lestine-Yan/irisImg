<template>
  <div class="relative min-h-screen">
    <LayoutAdminBackdrop />

    <div class="relative z-10 flex min-h-screen">
      <LayoutAppSidebar />

      <main class="min-w-0 flex-1">
        <div class="mx-auto max-w-6xl px-4 py-6 sm:px-6 lg:px-10 lg:py-10">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const { isAuthenticated, fetchMe } = useAuth()

// 后台布局挂载后统一刷新当前用户信息，供侧边栏 UserBadge 展示。
onMounted(() => {
  if (isAuthenticated.value) {
    fetchMe()
  }
})
</script>
