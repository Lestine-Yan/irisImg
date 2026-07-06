<template>
  <div class="flex min-h-screen bg-iris-cream">
    <LayoutAppSidebar />

    <main class="flex-1">
      <div class="mx-auto max-w-6xl px-6 py-8 lg:px-8">
        <slot />
      </div>
    </main>
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
