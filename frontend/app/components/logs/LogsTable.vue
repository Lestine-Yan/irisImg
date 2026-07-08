<template>
  <div>
    <!-- 加载态 -->
    <div
      v-if="loading"
      class="flex h-64 items-center justify-center rounded-2xl border border-dashed border-gray-200 bg-white/60"
    >
      <span class="text-sm text-gray-400">加载中…</span>
    </div>

    <!-- 错误态 -->
    <div
      v-else-if="error"
      class="flex h-64 flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-rose-200 bg-rose-50/50"
    >
      <p class="text-sm text-rose-600">{{ error }}</p>
      <button
        type="button"
        class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50"
        @click="emit('retry')"
      >
        重试
      </button>
    </div>

    <!-- 空态 -->
    <div
      v-else-if="!logs.length"
      class="flex h-64 flex-col items-center justify-center gap-1 rounded-2xl border border-dashed border-gray-300 bg-white/60"
    >
      <p class="text-sm text-gray-400">该筛选条件下暂无日志</p>
    </div>

    <!-- 列表 -->
    <div v-else class="overflow-x-auto rounded-2xl border border-gray-200 bg-white">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-100 bg-gray-50/60 text-left text-xs font-semibold uppercase tracking-wide text-gray-400">
            <th class="px-4 py-3">时间</th>
            <th class="px-4 py-3">级别</th>
            <th class="px-4 py-3">事件</th>
            <th class="px-4 py-3">方法 · 路径</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">耗时</th>
            <th class="px-4 py-3">来源</th>
            <th class="px-4 py-3">详情</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr
            v-for="log in logs"
            :key="log.id"
            class="transition-colors hover:bg-gray-50/60"
          >
            <!-- 时间 -->
            <td class="whitespace-nowrap px-4 py-3 text-gray-600">{{ formatDate(log.timestamp) }}</td>

            <!-- 级别徽章 -->
            <td class="px-4 py-3">
              <span class="rounded-full px-2 py-0.5 text-xs font-medium" :class="levelClass(log.level)">{{ log.level }}</span>
            </td>

            <!-- 事件 -->
            <td class="px-4 py-3">
              <code class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-600">{{ log.event }}</code>
            </td>

            <!-- 方法 · 路径 -->
            <td class="px-4 py-3">
              <div v-if="log.method" class="flex items-center gap-2">
                <code class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-600">{{ log.method }}</code>
                <span class="max-w-[220px] truncate text-gray-700" :title="log.path">{{ log.path }}</span>
              </div>
              <span v-else class="text-gray-300">-</span>
            </td>

            <!-- 状态 -->
            <td class="px-4 py-3 text-gray-600">{{ log.status ?? '-' }}</td>

            <!-- 耗时 -->
            <td class="px-4 py-3 text-gray-600">{{ log.duration_ms != null ? log.duration_ms + 'ms' : '-' }}</td>

            <!-- 来源 -->
            <td class="px-4 py-3 text-gray-600">
              <span v-if="log.username">{{ log.username }}</span>
              <span v-else-if="log.api_key_id != null">Key #{{ log.api_key_id }}</span>
              <span v-else class="text-gray-300">-</span>
            </td>

            <!-- 详情 -->
            <td class="max-w-[240px] truncate px-4 py-3 text-gray-500" :title="log.message">
              {{ log.message || '-' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { LogItem, LogLevel } from '~/composables/useLogs'
import { formatDate } from '~/composables/useImages'

defineProps<{
  logs: LogItem[]
  loading: boolean
  error: string | null
}>()
const emit = defineEmits<{ retry: [] }>()

// 级别徽章配色：error rose / warn amber / debug gray / info sky。
function levelClass(l: LogLevel): string {
  switch (l) {
    case 'error':
      return 'bg-rose-100 text-rose-700'
    case 'warn':
      return 'bg-amber-100 text-amber-700'
    case 'debug':
      return 'bg-gray-100 text-gray-600'
    default:
      return 'bg-sky-100 text-sky-700'
  }
}
</script>
