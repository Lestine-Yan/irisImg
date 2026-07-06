<template>
  <div class="flex items-center gap-3 rounded-xl px-2 py-1.5">
    <!-- 用户头像：白底圆 + 浅紫首字符 -->
    <div
      class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-gray-200 bg-white text-sm font-semibold text-iris-violet"
    >
      {{ initial }}
    </div>

    <div class="min-w-0 flex-1">
      <p class="truncate text-sm font-medium text-gray-900">{{ user?.username || '未登录' }}</p>
      <p class="truncate text-xs text-gray-500">已登录</p>
    </div>

    <button
      type="button"
      title="退出登录"
      class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-700"
      @click="handleLogout"
    >
      <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l3 3m0 0l-3 3m3-3H2.25"
        />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
const { user, logout } = useAuth()

// 用户名首字符大写作为头像文字；未登录时显示占位符。
const initial = computed(() => {
  const name = user.value?.username?.trim() ?? ''
  return name ? name.charAt(0).toUpperCase() : '?'
})

function handleLogout() {
  logout()
  navigateTo('/')
}
</script>
