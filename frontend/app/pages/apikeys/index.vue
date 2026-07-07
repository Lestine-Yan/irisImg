<template>
  <div class="space-y-6">
    <!-- 顶栏：标题 + 创建按钮（justify-between 布局） -->
    <div class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">APIkey 管理</h1>
        <p class="mt-1 text-sm text-gray-500">创建与维护 API 密钥</p>
      </div>
      <button
        type="button"
        class="inline-flex shrink-0 items-center gap-1.5 rounded-xl bg-iris-violet/10 px-3 py-2 text-sm font-medium text-iris-dark transition hover:bg-iris-violet/15"
        @click="dialogs.create = true"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4v16m8-8H4" />
        </svg>
        创建 Key
      </button>
    </div>

    <!-- 说明文本 -->
    <div class="flex items-start gap-2 rounded-xl bg-iris-violet/10 p-4 text-sm text-iris-dark">
      <svg class="mt-0.5 h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
      </svg>
      <p>APIkey 仅在创建时可见可复制，请妥善保存，不要与他人共享你的 APIkey，或将其暴露在浏览器或其他客户端代码中。</p>
    </div>

    <!-- 密钥列表 -->
    <ApikeysApiKeyTable
      :keys="keys"
      :loading="loading"
      :error="error"
      @rename="onRename"
      @reset="onReset"
      @revoke-delete="onRevokeDelete"
      @retry="fetchKeys"
    />

    <!-- 创建弹窗 -->
    <ApikeysCreateKeyDialog :open="dialogs.create" @close="dialogs.create = false" @created="onCreated" />

    <!-- 明文展示弹窗（创建 / 重置共用） -->
    <ApikeysPlaintextKeyDialog
      :open="dialogs.plaintext"
      :plaintext="plaintext"
      :title="plaintextTitle"
      @close="closePlaintext"
    />

    <!-- 重命名弹窗 -->
    <ApikeysRenameKeyDialog :open="dialogs.rename" :api-key="activeKey" @close="dialogs.rename = false" @renamed="onRenamed" />

    <!-- 重置弹窗 -->
    <ApikeysResetKeyDialog :open="dialogs.reset" :api-key="activeKey" @close="dialogs.reset = false" @reset="onResetDone" />

    <!-- 吊销 / 删除弹窗 -->
    <ApikeysRevokeDeleteDialog :open="dialogs.revokeDelete" :api-key="activeKey" @close="dialogs.revokeDelete = false" @done="onDestructiveDone" />
  </div>
</template>

<script setup lang="ts">
import type { APIKeyInfo, CreateAPIKeyResponse, ResetAPIKeyResponse } from '~/composables/useApiKeys'

definePageMeta({ middleware: 'auth' })

const { list } = useApiKeys()

const keys = ref<APIKeyInfo[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// 弹窗编排：同一时刻只开一个操作弹窗（明文展示可叠加在创建/重置之后）。
const dialogs = reactive({
  create: false,
  plaintext: false,
  rename: false,
  reset: false,
  revokeDelete: false,
})
const activeKey = ref<APIKeyInfo | null>(null)
const plaintext = ref('')
const plaintextTitle = ref('API Key 已创建')

async function fetchKeys() {
  loading.value = true
  error.value = null
  try {
    keys.value = await list()
  } catch (e) {
    error.value = e instanceof Error ? e.message : '加载密钥失败'
    keys.value = []
  } finally {
    loading.value = false
  }
}

// 创建成功 → 关闭创建弹窗 → 展示一次性明文 → 刷新列表。
function onCreated(resp: CreateAPIKeyResponse) {
  dialogs.create = false
  plaintext.value = resp.key
  plaintextTitle.value = 'API Key 已创建'
  dialogs.plaintext = true
  fetchKeys()
}

function closePlaintext() {
  dialogs.plaintext = false
  plaintext.value = ''
}

function onRename(k: APIKeyInfo) {
  activeKey.value = k
  dialogs.rename = true
}

function onRenamed() {
  dialogs.rename = false
  fetchKeys()
}

function onReset(k: APIKeyInfo) {
  activeKey.value = k
  dialogs.reset = true
}

function onResetDone(resp: ResetAPIKeyResponse) {
  dialogs.reset = false
  plaintext.value = resp.key
  plaintextTitle.value = 'API Key 已重置'
  dialogs.plaintext = true
  fetchKeys()
}

function onRevokeDelete(k: APIKeyInfo) {
  activeKey.value = k
  dialogs.revokeDelete = true
}

function onDestructiveDone(_mode: 'revoke' | 'purge') {
  dialogs.revokeDelete = false
  fetchKeys()
}

onMounted(fetchKeys)
</script>
