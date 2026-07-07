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
      v-else-if="!keys.length"
      class="flex h-64 flex-col items-center justify-center gap-1 rounded-2xl border border-dashed border-gray-300 bg-white/60"
    >
      <p class="text-sm text-gray-400">还没有 API Key，点击右上角「创建 Key」开始</p>
    </div>

    <!-- 列表 -->
    <div v-else class="overflow-x-auto rounded-2xl border border-gray-200 bg-white">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-100 bg-gray-50/60 text-left text-xs font-semibold uppercase tracking-wide text-gray-400">
            <th class="w-32 px-4 py-3">操作</th>
            <th class="px-4 py-3">名称</th>
            <th class="px-4 py-3">明文前缀</th>
            <th class="px-4 py-3">创建时间</th>
            <th class="px-4 py-3">最近使用</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <tr
            v-for="k in keys"
            :key="k.id"
            class="transition-colors hover:bg-gray-50/60"
            :class="k.revoked ? 'opacity-60' : ''"
          >
            <!-- 操作（最左）：SVG 图标按钮，title 即 hover 说明 -->
            <td class="px-4 py-3">
              <div class="flex items-center gap-1">
                <button
                  type="button"
                  title="重命名"
                  class="rounded-lg p-1.5 text-gray-400 transition hover:bg-iris-violet/15 hover:text-iris-dark"
                  @click="emit('rename', k)"
                >
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931z" />
                  </svg>
                </button>
                <button
                  type="button"
                  title="重置明文"
                  class="rounded-lg p-1.5 text-gray-400 transition hover:bg-iris-violet/15 hover:text-iris-dark"
                  @click="emit('reset', k)"
                >
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16.023 9.348h4.992V4.356M19.5 12c0 4.142-3.358 7.5-7.5 7.5S4.5 16.142 4.5 12 7.858 4.5 12 4.5c2.142 0 4.07.898 5.45 2.348" />
                  </svg>
                </button>
                <button
                  type="button"
                  title="吊销或删除"
                  class="rounded-lg p-1.5 text-gray-400 transition hover:bg-rose-100 hover:text-rose-600"
                  @click="emit('revokeDelete', k)"
                >
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                  </svg>
                </button>
              </div>
            </td>

            <!-- 名称 + 状态徽章 -->
            <td class="px-4 py-3">
              <div class="flex flex-wrap items-center gap-2">
                <span class="font-medium text-gray-900">{{ k.name }}</span>
                <span v-if="k.revoked" class="rounded-full bg-rose-100 px-2 py-0.5 text-xs text-rose-700">已吊销</span>
                <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500">{{ k.scope === 'readwrite' ? '读写' : '只读' }}</span>
              </div>
            </td>

            <!-- 明文前缀 -->
            <td class="px-4 py-3">
              <code class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-600">{{ k.prefix }}…</code>
            </td>

            <!-- 创建时间 -->
            <td class="px-4 py-3 text-gray-600">{{ formatDate(k.created_at) }}</td>

            <!-- 最近使用时间 -->
            <td class="px-4 py-3 text-gray-600">{{ k.last_used_at ? formatDate(k.last_used_at) : '从未' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { APIKeyInfo } from '~/composables/useApiKeys'

defineProps<{
  keys: APIKeyInfo[]
  loading: boolean
  error: string | null
}>()
const emit = defineEmits<{
  rename: [key: APIKeyInfo]
  reset: [key: APIKeyInfo]
  revokeDelete: [key: APIKeyInfo]
  retry: []
}>()
</script>
