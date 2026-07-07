<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/50" @click="close" />
      <div
        class="relative z-10 w-full max-w-lg overflow-hidden rounded-2xl bg-white shadow-xl"
        :class="contentClass"
      >
        <div
          v-if="title || $slots.header"
          class="flex items-center justify-between gap-4 border-b border-gray-100 px-6 py-4"
        >
          <h3 class="text-lg font-semibold text-gray-900">
            <slot name="header">{{ title }}</slot>
          </h3>
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
        <div class="px-6 py-5">
          <slot />
        </div>
        <div v-if="$slots.footer" class="flex justify-end gap-2 border-t border-gray-100 px-6 py-4">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
const props = defineProps<{
  open: boolean
  title?: string
  /** 透传给内容容器的 class，便于调整不同弹窗的宽度。 */
  contentClass?: string
}>()
const emit = defineEmits<{ close: [] }>()

function close() {
  emit('close')
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

// 弹窗显示时监听 ESC 关闭；隐藏或卸载时移除，避免泄漏。
watch(
  () => props.open,
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
