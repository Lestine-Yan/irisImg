<template>
  <div class="flex min-h-screen flex-col items-center justify-center bg-gray-50 p-6">
    <div class="w-full max-w-md rounded-2xl bg-white p-8 shadow-xl">
      <h1 class="mb-2 text-2xl font-bold text-gray-900">工作台</h1>
      <p class="mb-6 text-gray-600">欢迎回来，{{ user?.username || '...' }}</p>

      <button
        class="w-full rounded-xl bg-red-500 py-3 font-semibold text-white transition-colors hover:bg-red-600"
        @click="handleLogout"
      >
        退出登录
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const { user, isAuthenticated, fetchMe, logout } = useAuth()

onMounted(async () => {
  if (!isAuthenticated.value) {
    navigateTo('/')
    return
  }
  await fetchMe()
})

function handleLogout() {
  logout()
  navigateTo('/')
}
</script>
