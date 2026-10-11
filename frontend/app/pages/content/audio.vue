<template>
  <div class="space-y-6">
    <!-- 顶部标题 + 上传入口 -->
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">音频</h1>
        <p class="mt-1 text-sm text-gray-500">管理已上传的音频文件（MP3 / M4A / FLAC）</p>
      </div>
      <div class="flex shrink-0 items-center gap-2">
        <!-- 多选模式操作栏：删除（危险色，空选中禁用）+ 取消退出 -->
        <template v-if="multiSelect">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-xl bg-rose-50 px-3 py-2 text-sm font-medium text-rose-600 transition hover:bg-rose-100 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="!selectedIds.length"
            @click="showDeleteConfirm = true"
          >
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
            删除{{ selectedIds.length ? `（${selectedIds.length}）` : '' }}
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm font-medium text-gray-600 transition hover:bg-gray-50"
            @click="exitMultiSelect"
          >
            取消
          </button>
        </template>
        <!-- 默认态：多选入口 + 上传音频 -->
        <template v-else>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-xl bg-iris-violet/10 px-3 py-2 text-sm font-medium text-iris-dark transition hover:bg-iris-violet/15"
            @click="enterMultiSelect"
          >
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M8.25 3h7.5A5.25 5.25 0 0121 8.25v7.5A5.25 5.25 0 0115.75 21h-7.5A5.25 5.25 0 013 15.75v-7.5A5.25 5.25 0 018.25 3z"
              />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12.75l2.25 2.25L15 9.75" />
            </svg>
            多选
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-xl px-3 py-2 text-sm font-medium transition"
            :class="
              showUpload
                ? 'bg-iris-violet text-white shadow-sm'
                : 'bg-iris-violet/10 text-iris-dark hover:bg-iris-violet/15'
            "
            @click="showUpload = !showUpload"
          >
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M12 4v12m0-12l-4 4m4-4l4 4M4 20h16"
              />
            </svg>
            上传音频
          </button>
        </template>
      </div>
    </div>

    <!-- 上传栏：JWT 后台直传，不关联密钥；上传的音频只出现在「全部」里 -->
    <ContentAudioUploadPanel v-if="showUpload" @uploaded="onUploaded" />

    <div class="flex gap-6">
      <!-- 左侧：按 API Key 筛选的按钮栏（与图片页同款） -->
      <aside class="w-56 shrink-0">
        <div class="sticky top-8 space-y-1">
          <p class="px-3 pb-2 text-xs font-semibold uppercase tracking-wide text-gray-400">按 API Key 筛选</p>

          <!-- 全部按钮 -->
          <button
            type="button"
            class="flex w-full items-center justify-between rounded-xl px-3 py-2 text-sm font-medium transition"
            :class="
              selectedKeyId === null
                ? 'bg-iris-violet/15 text-iris-dark'
                : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
            "
            @click="selectKey(null)"
          >
            全部
            <span class="text-xs text-gray-400">{{ total }}</span>
          </button>

          <!-- 各 key.name 按钮 -->
          <button
            v-for="k in apiKeys"
            :key="k.id"
            type="button"
            class="flex w-full items-center justify-between gap-2 rounded-xl px-3 py-2 text-sm font-medium transition"
            :class="
              selectedKeyId === k.id
                ? 'bg-iris-violet/15 text-iris-dark'
                : k.revoked
                  ? 'text-gray-400 hover:bg-gray-100'
                  : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
            "
            @click="selectKey(k.id)"
          >
            <span class="truncate" :title="k.name">{{ k.name }}</span>
            <span v-if="k.revoked" class="shrink-0 text-xs text-gray-300">已吊销</span>
          </button>

          <p v-if="!apiKeys.length && !keysLoading" class="px-3 py-2 text-xs text-gray-400">暂无密钥</p>
        </div>
      </aside>

      <!-- 右侧：音频列表 -->
      <div class="min-w-0 flex-1">
        <!-- 状态行 + 分页 -->
        <div class="mb-4 flex items-center justify-between gap-4">
          <p class="text-sm text-gray-500">
            共 <span class="font-medium text-gray-700">{{ total }}</span> 个
            <span v-if="selectedKeyId !== null">· 已按 Key 筛选</span>
            <span v-if="multiSelect && selectedIds.length" class="text-iris-dark">· 已选 {{ selectedIds.length }} 个</span>
          </p>
          <UiPagination :page="page" :total-pages="totalPages" :loading="loading" @change="goPage" />
        </div>

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
            class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 hover:bg-gray-50"
            @click="fetchAudios"
          >
            重试
          </button>
        </div>

        <!-- 空态 -->
        <div
          v-else-if="!audios.length"
          class="flex h-64 flex-col items-center justify-center gap-1 rounded-2xl border border-dashed border-gray-300 bg-white/60"
        >
          <p class="text-sm text-gray-400">该筛选条件下暂无音频</p>
        </div>

        <!-- 音频列表：纵向堆叠的横向条卡（左封面、右简介） -->
        <div v-else class="space-y-3">
          <ContentAudioCard
            v-for="a in audios"
            :key="a.id"
            :audio="a"
            :selectable="multiSelect"
            :selected="selectedIds.includes(a.id)"
            @click="onCardClick"
          />
        </div>
      </div>
    </div>

    <!-- 详情弹窗（含封面 URL 编辑） -->
    <ContentAudioDetailDialog
      :audio="selectedAudio"
      :key-name="selectedKeyName"
      @close="closeDetail"
      @updated="onAudioUpdated"
    />

    <!-- 批量删除确认弹窗（账号密码二次确认） -->
    <ContentAudioBatchDeleteDialog :open="showDeleteConfirm" :ids="selectedIds" @close="showDeleteConfirm = false" @done="onDeleted" />
  </div>
</template>

<script setup lang="ts">
import type { AudioItem, BatchDeleteAudiosResult } from '~/composables/useAudios'

definePageMeta({ middleware: 'auth' })

interface ApiKeyOption {
  id: number
  name: string
  revoked?: boolean
}

const PAGE_SIZE = 24
const { list } = useAudios()
const { get } = useApi()

// 筛选与分页状态
const apiKeys = ref<ApiKeyOption[]>([])
const keysLoading = ref(false)
const selectedKeyId = ref<number | null>(null)

const audios = ref<AudioItem[]>([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const error = ref<string | null>(null)

// 详情弹窗状态
const selectedAudio = ref<AudioItem | null>(null)

// 上传栏显隐
const showUpload = ref(false)

// 多选模式状态
const multiSelect = ref(false)
// 选中的音频 ID（跨页累计；切换 Key 筛选会退出多选并清空）
const selectedIds = ref<number[]>([])
// 批量删除确认弹窗显隐
const showDeleteConfirm = ref(false)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / PAGE_SIZE)))

// id → name 映射，供详情弹窗解析来源 Key 名称（无需后端联表）。
const keyNameById = computed(() => {
  const m = new Map<number, string>()
  for (const k of apiKeys.value) m.set(k.id, k.name)
  return m
})

const selectedKeyName = computed(() => {
  const id = selectedAudio.value?.key_id
  // key_id 为空代表后台 JWT 直传（admin），详情里来源展示为 admin。
  if (id == null) return 'admin'
  return keyNameById.value.get(id) || `#${id}`
})

async function fetchApiKeys() {
  keysLoading.value = true
  try {
    const data = await get<{ items: ApiKeyOption[] }>('/apikeys')
    apiKeys.value = data.items ?? []
  } catch {
    apiKeys.value = []
  } finally {
    keysLoading.value = false
  }
}

async function fetchAudios() {
  loading.value = true
  error.value = null
  try {
    const data = await list({
      keyId: selectedKeyId.value ?? undefined,
      order: 'asc',
      page: page.value,
      pageSize: PAGE_SIZE,
    })
    audios.value = data.items ?? []
    total.value = data.total ?? 0
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载音频失败'
    audios.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 切换筛选 Key 时重置到第 1 页再拉取。
function selectKey(id: number | null) {
  if (selectedKeyId.value === id) return
  selectedKeyId.value = id
  page.value = 1
  // 跨筛选维度保留选中违背直觉、误删风险高，直接退出多选。
  if (multiSelect.value) exitMultiSelect()
  fetchAudios()
}

function goPage(p: number) {
  if (p < 1 || p > totalPages.value || p === page.value) return
  page.value = p
  fetchAudios()
}

function openDetail(a: AudioItem) {
  selectedAudio.value = a
}

function closeDetail() {
  selectedAudio.value = null
}

// 详情弹窗里改封面成功：就地替换列表项并同步（保持打开的详情数据）最新，无需整页刷新。
function onAudioUpdated(updated: AudioItem) {
  const idx = audios.value.findIndex((a) => a.id === updated.id)
  if (idx >= 0) audios.value[idx] = updated
  selectedAudio.value = updated
}

// 进入多选模式：清空选中并收起上传面板，避免两种模式叠屏。
function enterMultiSelect() {
  multiSelect.value = true
  selectedIds.value = []
  showUpload.value = false
}

// 退出多选模式：清空选中与确认弹窗。
function exitMultiSelect() {
  multiSelect.value = false
  selectedIds.value = []
  showDeleteConfirm.value = false
}

// 卡片点击语义分叉：多选模式切换选中（不打开详情），默认模式打开详情弹窗。
// 封面上的播放/暂停按钮在组件内部已 @click.stop，不会触发本回调。
function onCardClick(a: AudioItem) {
  if (multiSelect.value) toggleSelect(a)
  else openDetail(a)
}

// 切换一个音频的选中状态：再点一次取消选中。
function toggleSelect(a: AudioItem) {
  const i = selectedIds.value.indexOf(a.id)
  if (i >= 0) selectedIds.value.splice(i, 1)
  else selectedIds.value.push(a.id)
}

// 批量删除完成：清空选中、按删除后的总数收敛页码（当前页被删空时回退）再刷新。
// 保留多选模式，便于连续分批删除；点「取消」才退出。
async function onDeleted(result: BatchDeleteAudiosResult) {
  showDeleteConfirm.value = false
  selectedIds.value = []
  const remaining = total.value - result.deleted
  const maxPage = Math.max(1, Math.ceil(remaining / PAGE_SIZE))
  if (page.value > maxPage) page.value = maxPage
  await fetchAudios()
}

// 上传成功后刷新列表。admin 直传的音频 key_id 为空，只在不按密钥过滤时可见：
// 若当前在按密钥筛选，先切回「全部」（会重置到第 1 页并拉取），否则直接刷新当前页。
function onUploaded(_audio: AudioItem) {
  if (selectedKeyId.value !== null) {
    selectKey(null)
  } else {
    fetchAudios()
  }
}

onMounted(() => {
  // 密钥列表与音频列表互不依赖，并行拉取。
  fetchApiKeys()
  fetchAudios()
})
</script>
