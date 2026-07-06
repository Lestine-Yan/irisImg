<template>
  <Teleport to="body">
    <div v-if="image" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/50" @click="close" />
      <div class="relative z-10 flex max-h-[90vh] w-full max-w-4xl overflow-hidden rounded-2xl bg-white shadow-xl">
        <!-- 左：大图预览 -->
        <div class="flex w-1/2 items-center justify-center bg-gray-900">
          <img
            v-if="resolvedUrl"
            :src="resolvedUrl"
            :alt="image.filename"
            class="max-h-[90vh] w-full object-contain"
          />
        </div>
        <!-- 右：元数据 -->
        <div class="flex w-1/2 flex-col overflow-y-auto p-6">
          <div class="flex items-start justify-between gap-4">
            <h3 class="break-all text-lg font-semibold text-gray-900">{{ image.filename }}</h3>
            <button
              type="button"
              class="shrink-0 rounded-lg p-1 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700"
              @click="close"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <dl class="mt-4 space-y-3 text-sm">
            <div v-for="row in rows" :key="row.label" class="flex gap-3">
              <dt class="w-20 shrink-0 text-gray-400">{{ row.label }}</dt>
              <dd class="flex-1 break-all text-gray-800">
                <code v-if="row.mono" class="text-xs text-gray-500">{{ row.value }}</code>
                <template v-else>{{ row.value }}</template>
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { ImageItem } from '~/composables/useImages'

const props = defineProps<{ image: ImageItem | null; keyName?: string }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const resolvedUrl = computed(() => resolveImageUrl(props.image?.url))

const rows = computed(() => {
  const img = props.image
  if (!img) return []
  return [
    { label: '来源 Key', value: props.keyName || '—' },
    {
      label: '尺寸',
      value: img.width && img.height ? `${img.width} × ${img.height}` : '—',
    },
    { label: '大小', value: formatBytes(img.size) },
    { label: '类型', value: img.mime_type || '—' },
    { label: '上传时间', value: formatDate(img.created_at) },
    { label: '文件 ID', value: String(img.id) },
    { label: '哈希', value: img.hash, mono: true },
    { label: '访问 URL', value: img.url, mono: true },
  ]
})

function close() {
  emit('close')
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

// 弹窗显示时监听 ESC 关闭；隐藏或卸载时移除，避免泄漏。
watch(
  () => props.image,
  (v) => {
    if (import.meta.client) {
      if (v) window.addEventListener('keydown', onKey)
      else window.removeEventListener('keydown', onKey)
    }
  },
)
onBeforeUnmount(() => {
  if (import.meta.client) window.removeEventListener('keydown', onKey)
})
</script>
