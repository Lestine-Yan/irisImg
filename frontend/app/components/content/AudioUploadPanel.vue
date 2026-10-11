<template>
  <div class="space-y-3">
    <!-- 拖拽 / 点击选择区 -->
    <div
      class="flex cursor-pointer flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed px-6 py-8 text-center transition"
      :class="
        dragging
          ? 'border-iris-violet bg-iris-violet/5'
          : 'border-gray-300 bg-white/60 hover:border-iris-violet/50 hover:bg-gray-50'
      "
      @click="triggerPick"
      @dragenter.prevent="onDragEnter"
      @dragover.prevent="onDragOver"
      @dragleave.prevent="onDragLeave"
      @drop.prevent="onDrop"
    >
      <svg class="h-8 w-8 text-iris-violet/70" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M9 18V5l12-2v13M9 18a3 3 0 11-6 0 3 3 0 016 0M21 16a3 3 0 11-6 0 3 3 0 016 0"
        />
      </svg>
      <p class="text-sm text-gray-600">
        拖拽音频到此处，或<span class="text-iris-dark underline">点击选择</span>
      </p>
      <p class="text-xs text-gray-400">支持 MP3 / M4A / FLAC，多选可批量上传</p>
      <input ref="inputRef" type="file" accept=".mp3,.m4a,.flac,audio/*" multiple class="hidden" @change="onPick" />
    </div>

    <!-- 上传队列 -->
    <ul v-if="queue.length" class="space-y-2">
      <li
        v-for="item in queue"
        :key="item.id"
        class="flex items-center gap-3 rounded-xl border border-gray-200 bg-white px-3 py-2"
      >
        <span class="h-2 w-2 shrink-0 rounded-full" :class="dotClass(item.status)" />
        <div class="min-w-0 flex-1">
          <p class="truncate text-sm text-gray-800" :title="item.file.name">{{ item.file.name }}</p>
          <p class="text-xs text-gray-400">{{ formatBytes(item.file.size) }}</p>
        </div>
        <div class="shrink-0 text-right">
          <p v-if="item.status === 'uploading'" class="text-xs text-iris-dark">上传中…</p>
          <p v-else-if="item.status === 'success'" class="text-xs text-emerald-600">已完成</p>
          <p v-else-if="item.status === 'error'" class="max-w-[12rem] truncate text-xs text-rose-600" :title="item.error">
            {{ item.error }}
          </p>
          <p v-else class="text-xs text-gray-400">等待中</p>
        </div>
        <button
          v-if="item.status !== 'uploading'"
          type="button"
          class="shrink-0 rounded p-1 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700"
          title="移除"
          @click="removeItem(item.id)"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </li>
    </ul>

    <div v-if="hasCompleted" class="flex justify-end">
      <button type="button" class="text-xs text-gray-500 transition hover:text-gray-800" @click="clearCompleted">
        清空已完成
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AudioItem } from '~/composables/useAudios'
import { formatBytes } from '~/composables/useImages'

const emit = defineEmits<{ (e: 'uploaded', audio: AudioItem): void }>()

type Status = 'pending' | 'uploading' | 'success' | 'error'
interface QueueItem {
  id: number
  file: File
  status: Status
  error?: string
  result?: AudioItem
}

const { upload } = useAudios()

const inputRef = ref<HTMLInputElement | null>(null)
const queue = ref<QueueItem[]>([])
const dragging = ref(false)
// 用深度计数避免子元素反复触发 dragenter/leave 导致高亮闪烁。
const dragDepth = ref(0)
const processing = ref(false)

// 队列项自增 id（组件实例内有效，无需全局唯一）。
let nextId = 0

const hasCompleted = computed(() =>
  queue.value.some((i) => i.status === 'success' || i.status === 'error'),
)

function dotClass(status: Status): string {
  switch (status) {
    case 'success':
      return 'bg-emerald-500'
    case 'error':
      return 'bg-rose-500'
    case 'uploading':
      return 'bg-iris-violet animate-pulse'
    default:
      return 'bg-gray-300'
  }
}

function triggerPick() {
  inputRef.value?.click()
}

function onPick(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files?.length) addFiles(target.files)
  // 重置 value，允许再次选择同一文件触发 change。
  target.value = ''
}

// isAudioFile 判定接受范围：扩展名 mp3/m4a/flac 或 audio/* MIME。
// 部分操作系统的 <input> 会把 .flac 的 MIME 报成空串，故不能只信 f.type。
function isAudioFile(f: File): boolean {
  if (f.type.startsWith('audio/')) return true
  const ext = f.name.slice(f.name.lastIndexOf('.') + 1).toLowerCase()
  return ext === 'mp3' || ext === 'm4a' || ext === 'flac'
}

function addFiles(files: FileList | File[]) {
  // 非音频静默忽略（input accept 已限，拖拽时再兜底过滤一次）。
  const audios = Array.from(files).filter(isAudioFile)
  for (const f of audios) {
    queue.value.push({ id: ++nextId, file: f, status: 'pending' })
  }
  void processQueue()
}

function onDragEnter(e: DragEvent) {
  e.preventDefault()
  dragDepth.value++
  dragging.value = true
}

function onDragOver(e: DragEvent) {
  // dragover 必须 preventDefault，否则 drop 不会触发。
  e.preventDefault()
}

function onDragLeave(e: DragEvent) {
  e.preventDefault()
  dragDepth.value = Math.max(0, dragDepth.value - 1)
  if (dragDepth.value === 0) dragging.value = false
}

function onDrop(e: DragEvent) {
  e.preventDefault()
  dragDepth.value = 0
  dragging.value = false
  if (e.dataTransfer?.files?.length) addFiles(e.dataTransfer.files)
}

// processQueue 顺序上传（一次一个），避免并发冲击；每个成功即 emit 让父组件刷新。
async function processQueue() {
  if (processing.value) return
  processing.value = true
  try {
    while (true) {
      const item = queue.value.find((i) => i.status === 'pending')
      if (!item) break
      item.status = 'uploading'
      try {
        const audio = await upload(item.file)
        item.status = 'success'
        item.result = audio
        emit('uploaded', audio)
      } catch (e) {
        item.status = 'error'
        item.error = resolveErrorMsg(e)
      }
    }
  } finally {
    processing.value = false
  }
}

// resolveErrorMsg 优先取后端返回的业务 message（ofetch 把响应体挂在 err.data 上）。
function resolveErrorMsg(e: unknown): string {
  if (e && typeof e === 'object') {
    const any = e as { data?: { message?: string }; message?: string }
    if (any.data?.message) return any.data.message
    if (any.message) return any.message
  }
  return '上传失败'
}

function removeItem(id: number) {
  const idx = queue.value.findIndex((i) => i.id === id)
  if (idx >= 0 && queue.value[idx].status !== 'uploading') {
    queue.value.splice(idx, 1)
  }
}

function clearCompleted() {
  queue.value = queue.value.filter((i) => i.status === 'pending' || i.status === 'uploading')
}
</script>
