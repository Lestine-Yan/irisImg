<template>
  <div
    class="flex flex-col items-center gap-2 rounded-2xl px-1 py-1 xl:flex-row xl:items-center xl:gap-3 xl:border xl:border-white/70 xl:bg-white/45 xl:p-2.5"
  >
    <!-- 用户头像：浅紫渐变圆 + 首字符 -->
    <div
      class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-white/80 bg-gradient-to-br from-white to-[#F0EBFC] text-[13px] font-semibold text-iris-dark shadow-sm"
    >
      {{ initial }}
    </div>

    <div class="hidden min-w-0 flex-1 xl:block">
      <p class="truncate text-[12.5px] font-semibold text-[#1C1A27]">{{ user?.username || '未登录' }}</p>
      <p class="truncate text-[10px] text-[#8A8696]">已登录 · 私人空间</p>
    </div>

    <button
      type="button"
      title="退出登录"
      aria-label="退出登录"
      class="rounded-lg p-1.5 text-[#8A8696] transition-colors hover:bg-white/60 hover:text-[#1C1A27]"
      @click="handleLogout"
    >
      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
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
