<template>
  <div
    class="group flex cursor-pointer items-center gap-4 rounded-2xl border bg-white px-4 py-3 text-left transition focus:outline-none focus:ring-2 focus:ring-iris-violet/40"
    :class="borderClass"
    :aria-pressed="selectable ? selected : undefined"
    role="button"
    tabindex="0"
    @click="emit('click', audio)"
    @keydown.enter.prevent="emit('click', audio)"
  >
    <!-- 左：方形封面（内嵌/覆盖封面，无则占位音符）+ 播放/暂停按钮 + 多选徽标 -->
    <div class="relative h-20 w-20 shrink-0 overflow-hidden rounded-xl bg-gradient-to-br from-iris-violet/25 to-iris-sky/25">
      <img
        v-if="coverSrc"
        :src="coverSrc"
        :alt="audio.title || audio.filename"
        loading="lazy"
        class="h-full w-full object-cover transition duration-200 group-hover:scale-[1.03]"
      />
      <svg
        v-else
        class="absolute inset-0 m-auto h-8 w-8 text-iris-violet/60"
        fill="none"
        stroke="currentColor"
        stroke-width="1.5"
        viewBox="0 0 24 24"
      >
        <!-- 音符（feather music）：竖线 + 符头双圆 -->
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 18V5l12-2v13" />
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 18a3 3 0 11-6 0 3 3 0 016 0" />
        <path stroke-linecap="round" stroke-linejoin="round" d="M21 16a3 3 0 11-6 0 3 3 0 016 0" />
      </svg>

      <!-- 播放 / 暂停：默认 hover 才显形，播放中常显；click.stop 避免触发卡片的选中/详情 -->
      <button
        type="button"
        class="absolute inset-0 m-auto flex h-9 w-9 items-center justify-center rounded-full bg-black/45 text-white opacity-0 backdrop-blur-sm transition group-hover:opacity-100 focus:opacity-100"
        :class="{ 'opacity-100': isPlaying(audio.id) }"
        :title="isPlaying(audio.id) ? '暂停' : '播放'"
        :aria-label="isPlaying(audio.id) ? '暂停' : '播放'"
        @click.stop="toggle(audio.id, resolvedAudioUrl)"
      >
        <svg v-if="isPlaying(audio.id)" class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
          <path d="M7 5h3.2v14H7zM13.8 5H17v14h-3.2z" />
        </svg>
        <svg v-else class="h-4 w-4 translate-x-[1px]" fill="currentColor" viewBox="0 0 24 24">
          <path d="M7 4.5l13 7.5-13 7.5v-15z" />
        </svg>
      </button>

      <!-- 多选模式：右上角勾选徽标，两态切换（与 ImageCard 的右下角徽标同款交互） -->
      <span
        v-if="selectable"
        class="absolute -right-1.5 -top-1.5 flex h-6 w-6 items-center justify-center rounded-full border transition"
        :class="
          selected
            ? 'border-iris-dark bg-iris-dark text-white shadow-sm'
            : 'border-white/70 bg-black/30 text-transparent backdrop-blur-sm'
        "
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M4.5 12.75l6 6 9-13.5" />
        </svg>
      </span>
    </div>

    <!-- 右：简介数据（标题 / 艺术家—专辑 / 时长·格式·大小·日期） -->
    <div class="min-w-0 flex-1">
      <p class="truncate text-sm font-medium text-gray-800" :title="audio.title || audio.filename">
        {{ audio.title || audio.filename }}
      </p>
      <p class="mt-0.5 truncate text-xs text-gray-500" :title="artistAlbum">{{ artistAlbum || '—' }}</p>
      <p class="mt-1 truncate text-xs text-gray-400">
        {{ formatDuration(audio.duration_seconds) }} · {{ formatLabel }} · {{ formatBytes(audio.size) }} ·
        {{ formatDate(audio.created_at) }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AudioItem } from '~/composables/useAudios'
import { effectiveCover, formatDuration, resolveAudioUrl } from '~/composables/useAudios'
import { formatBytes, formatDate } from '~/composables/useImages'

const props = defineProps<{
  audio: AudioItem
  /** 是否处于多选模式：显示勾选徽标，点击卡片切换选中。 */
  selectable?: boolean
  /** 多选模式下该卡片是否被选中：主色描边 + 实底勾。 */
  selected?: boolean
}>()
const emit = defineEmits<{ (e: 'click', audio: AudioItem): void }>()

// 全局单实例播放器：currentId/playing 跨卡片共享，同时至多一个音频在播。
const { toggle, isPlaying } = useAudioPlayer()

const coverSrc = computed(() => {
  const url = effectiveCover(props.audio)
  return url ? resolveAudioUrl(url) : ''
})
const resolvedAudioUrl = computed(() => resolveAudioUrl(props.audio.url))

// 「艺术家 — 专辑」，两者都缺时整行由模板降级为占位横线。
const artistAlbum = computed(() => {
  const a = props.audio.artist || ''
  const b = props.audio.album || ''
  if (a && b) return `${a} — ${b}`
  return a || b
})

const formatLabel = computed(() => (props.audio.format || '?').toUpperCase())

// 边框样式三态互斥：选中 > 多选未选中 > 默认（与 ImageCard 同款规则）。
const borderClass = computed(() => {
  if (props.selected) return 'border-iris-dark ring-2 ring-iris-violet/30'
  if (props.selectable) return 'border-gray-200 hover:border-iris-dark/50'
  return 'border-gray-200 hover:border-iris-violet/60 hover:shadow-sm'
})
</script>
