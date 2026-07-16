<template>
  <div class="rounded-2xl border border-gray-200 bg-white p-6">
    <div class="flex items-start justify-between gap-4">
      <div class="min-w-0">
        <p class="text-sm font-medium text-gray-500">{{ label }}</p>
        <p class="mt-2 text-3xl font-bold text-gray-900">{{ value }}</p>
        <p v-if="hint" class="mt-1 text-xs text-gray-400">{{ hint }}</p>
      </div>
      <span
        v-if="icon"
        class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-iris-violet/10 text-iris-dark"
      >
        <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" :d="icon" />
        </svg>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * StatCard 是仪表盘统计卡片：左侧标题 + 大数字 + 可选副标题，右侧 iris 图标徽章。
 *
 * 容器样式沿用 ConfigSection 的 rounded-2xl border bg-white p-6 范式；
 * 图标徽章用 iris-violet/10 底 + iris-dark 字色，与侧边栏激活态呼应。
 * value 由父组件格式化后传入（如存储大小用 formatBytes）。
 */
defineProps<{
  label: string
  value: string | number
  icon?: string
  hint?: string
}>()
</script>
