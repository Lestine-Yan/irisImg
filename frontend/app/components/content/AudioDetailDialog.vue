<template>
  <Teleport to="body">
    <div v-if="audio" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/50" @click="close" />
      <div class="relative z-10 flex max-h-[90vh] w-full max-w-4xl overflow-hidden rounded-2xl bg-white shadow-xl">
        <!-- 左：生效封面大图（cover_url 覆盖 > 内嵌封面 > 占位音符） -->
        <div class="flex w-2/5 items-center justify-center bg-gray-900">
          <img
            v-if="coverSrc"
            :src="coverSrc"
            :alt="audio.title || audio.filename"
            class="max-h-[90vh] w-full object-cover"
          />
          <svg
            v-else
            class="h-16 w-16 text-iris-violet/50"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 18V5l12-2v13" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 18a3 3 0 11-6 0 3 3 0 016 0" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 16a3 3 0 11-6 0 3 3 0 016 0" />
          </svg>
        </div>
        <!-- 右：元数据 + 封面 URL 编辑 -->
        <div class="flex w-3/5 flex-col overflow-y-auto p-6">
          <div class="flex items-start justify-between gap-4">
            <h3 class="break-all text-lg font-semibold text-gray-900" :title="audio.filename">
              {{ audio.title || audio.filename }}
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
          <dl class="mt-4 space-y-3 text-sm">
            <div v-for="row in rows" :key="row.label" class="flex gap-3">
              <dt class="w-24 shrink-0 text-gray-400">{{ row.label }}</dt>
              <dd class="min-w-0 flex-1 break-all text-gray-800">
                <code v-if="row.mono" class="break-all text-xs text-gray-500">{{ row.value }}</code>
                <template v-else>{{ row.value }}</template>
              </dd>
            </div>
          </dl>

          <!-- 封面 URL 编辑：输入的是「用户覆盖值」，留空表示用内嵌封面 -->
          <form class="mt-6 space-y-3 border-t border-gray-100 pt-5" @submit.prevent="handleSave">
            <div>
              <label for="audio-cover-url" class="block text-sm font-medium text-gray-700">封面 URL</label>
              <input
                id="audio-cover-url"
                v-model="form.coverUrl"
                type="text"
                placeholder="留空使用内嵌封面"
                class="mt-1 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-gray-900 outline-none transition-colors focus:border-iris-dark focus:bg-white focus:ring-2 focus:ring-iris-dark/10"
                :class="{ 'border-red-300 focus:border-red-500 focus:ring-red-500/10': errors.coverUrl }"
                @input="errors.coverUrl = ''"
              />
              <p v-if="errors.coverUrl" class="mt-1.5 text-sm text-red-500">{{ errors.coverUrl }}</p>
              <p class="mt-1.5 text-xs text-gray-400">填入 http(s):// 或 / 开头的图片地址覆盖默认封面，留空恢复内嵌封面</p>
            </div>
            <div v-if="serverError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600">{{ serverError }}</div>
            <div class="flex items-center justify-end gap-2">
              <button
                type="button"
                class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50"
                @click="close"
              >
                取消
              </button>
              <button
                type="button"
                :disabled="loading"
                class="inline-flex items-center justify-center rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-iris-dark transition hover:bg-iris-violet/10 disabled:cursor-not-allowed disabled:opacity-60"
                @click="handleReset"
              >
                恢复默认
              </button>
              <button
                type="submit"
                :disabled="loading"
                class="inline-flex items-center justify-center gap-1.5 rounded-xl bg-iris-violet/10 px-3 py-2 text-sm font-medium text-iris-dark transition hover:bg-iris-violet/15 disabled:cursor-not-allowed disabled:opacity-60"
              >
                <span v-if="loading" class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-iris-violet/20 border-t-iris-dark"></span>
                保存
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { AudioItem } from '~/composables/useAudios'
import { effectiveCover, formatDuration, resolveAudioUrl } from '~/composables/useAudios'
import { formatBytes, formatDate } from '~/composables/useImages'

const props = defineProps<{ audio: AudioItem | null; keyName?: string }>()
const emit = defineEmits<{
  close: []
  /** 封面修改成功，返回更新后的 AudioItem，页面就地替换列表项（无需整页刷新）。 */
  updated: [audio: AudioItem]
}>()

const { updateCover } = useAudios()

const form = reactive({ coverUrl: '' })
const errors = reactive({ coverUrl: '' })
const serverError = ref('')
const loading = ref(false)

const coverSrc = computed(() => {
  const url = props.audio ? effectiveCover(props.audio) : ''
  return url ? resolveAudioUrl(url) : ''
})

// 声道数的可读表达。
function channelLabel(n: number): string {
  if (!n) return '—'
  if (n === 1) return '单声道'
  if (n === 2) return '双声道'
  return `${n} 声道`
}

const rows = computed(() => {
  const a = props.audio
  if (!a) return []
  return [
    { label: '来源 Key', value: props.keyName || '—' },
    { label: '标题', value: a.title || a.filename },
    { label: '艺术家', value: a.artist || '—' },
    { label: '专辑', value: a.album || '—' },
    { label: '专辑艺术家', value: a.album_artist || '—' },
    { label: '作曲', value: a.composer || '—' },
    { label: '流派', value: a.genre || '—' },
    {
      label: '年份 / 音轨',
      value: [a.year || null, a.track_number || null].some(Boolean)
        ? [a.year || '—', a.track_number || '—'].join(' · ')
        : '—',
    },
    { label: '碟号', value: a.disc_number ? String(a.disc_number) : '—' },
    { label: '时长', value: formatDuration(a.duration_seconds) },
    { label: '码率', value: a.bitrate_kbps ? `${a.bitrate_kbps} kbps` : '—' },
    { label: '采样率', value: a.sample_rate ? `${a.sample_rate} Hz` : '—' },
    { label: '声道', value: channelLabel(a.channels) },
    { label: '位深', value: a.bit_depth ? `${a.bit_depth} bit` : '—' },
    { label: '格式 / 大小', value: `${(a.format || '—').toUpperCase()} · ${formatBytes(a.size)}` },
    { label: 'MIME', value: a.mime_type || '—' },
    { label: '上传时间', value: formatDate(a.created_at) },
    { label: '文件 ID', value: String(a.id) },
    { label: '哈希', value: a.hash, mono: true },
    { label: '访问 URL', value: a.url, mono: true },
    { label: '内嵌封面', value: a.embedded_cover_url || '无', mono: true },
  ]
})

// 弹窗每次打开：重置表单为「当前覆盖值」（而非生效封面），错误清空。
// 输入框留空 = 清除覆盖，恢复内嵌封面——与后端「空串清除」语义一致。
watch(
  () => props.audio,
  (a) => {
    if (a) {
      form.coverUrl = a.cover_url ?? ''
      errors.coverUrl = ''
      serverError.value = ''
      loading.value = false
    }
  },
)

// validate 轻校验：空串合法（恢复默认）；非空须以 http(s):// 或 / 开头且 ≤ 2048 字符
// （长度上限与后端 binding max=2048 对齐）。
function validate(): boolean {
  const v = form.coverUrl.trim()
  if (!v) return true
  if (v.length > 2048) {
    errors.coverUrl = '封面 URL 过长（上限 2048 字符）'
    return false
  }
  if (!/^(https?:\/\/|\/)/i.test(v)) {
    errors.coverUrl = '封面 URL 需以 http(s):// 或 / 开头'
    return false
  }
  return true
}

// handleSave 提交当前输入框内容（空串 = 清除覆盖）。
async function handleSave() {
  if (!props.audio || loading.value) return
  if (!validate()) return
  await submit(form.coverUrl.trim())
}

// handleReset 恢复默认：清除用户覆盖，回到内嵌封面。
async function handleReset() {
  if (!props.audio || loading.value) return
  form.coverUrl = ''
  errors.coverUrl = ''
  await submit('')
}

async function submit(coverUrl: string) {
  if (!props.audio) return
  loading.value = true
  serverError.value = ''
  try {
    const updated = await updateCover(props.audio.id, coverUrl)
    emit('updated', updated)
  } catch (err: unknown) {
    serverError.value = err instanceof Error ? err.message : '修改封面失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function close() {
  emit('close')
}

function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') close()
}

// 弹窗显示时监听 ESC 关闭；隐藏或卸载时移除，避免泄漏。
watch(
  () => props.audio,
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
