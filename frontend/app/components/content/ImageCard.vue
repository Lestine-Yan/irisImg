<template>
  <button
    type="button"
    class="group flex flex-col overflow-hidden rounded-2xl border border-gray-200 bg-white text-left transition hover:border-iris-violet/60 hover:shadow-sm focus:outline-none focus:ring-2 focus:ring-iris-violet/40"
    @click="emit('click', image)"
  >
    <div class="relative aspect-square w-full overflow-hidden bg-gray-50">
      <img
        :src="resolvedUrl"
        :alt="image.filename"
        loading="lazy"
        class="h-full w-full object-cover transition duration-200 group-hover:scale-[1.03]"
      />
    </div>
    <div class="flex flex-col gap-0.5 px-3 py-2">
      <span class="truncate text-sm font-medium text-gray-800" :title="image.filename">{{ image.filename }}</span>
      <span class="text-xs text-gray-400">{{ formatBytes(image.size) }} · {{ formatDate(image.created_at) }}</span>
    </div>
  </button>
</template>

<script setup lang="ts">
import type { ImageItem } from '~/composables/useImages'

const props = defineProps<{ image: ImageItem }>()
const emit = defineEmits<{ (e: 'click', image: ImageItem): void }>()

const resolvedUrl = computed(() => resolveImageUrl(props.image.url))
</script>
