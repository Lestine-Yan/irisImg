<template>
  <UiBaseDialog :open="open" title="吊销或删除 API Key" content-class="max-w-xl" @close="emit('close')">
    <div class="space-y-5">
      <!-- 操作模式选择 -->
      <div class="space-y-2">
        <span class="block text-sm font-medium text-gray-700">选择操作</span>
        <button
          type="button"
          class="flex w-full items-start gap-3 rounded-xl border px-4 py-3 text-left transition"
          :class="form.mode === 'revoke' ? 'border-iris-dark bg-iris-violet/5' : 'border-gray-200 hover:bg-gray-50'"
          @click="form.mode = 'revoke'"
        >
          <span class="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border" :class="form.mode === 'revoke' ? 'border-iris-dark' : 'border-gray-300'">
            <span v-if="form.mode === 'revoke'" class="h-2 w-2 rounded-full bg-iris-dark"></span>
          </span>
          <span>
            <span class="block text-sm font-medium text-gray-900">仅吊销</span>
            <span class="mt-0.5 block text-xs text-gray-500">密钥保留但无法通过鉴权，可随时重置或删除。</span>
          </span>
        </button>
        <button
          type="button"
          class="flex w-full items-start gap-3 rounded-xl border px-4 py-3 text-left transition"
          :class="form.mode === 'purge' ? 'border-rose-500 bg-rose-50' : 'border-gray-200 hover:bg-gray-50'"
          @click="form.mode = 'purge'"
        >
          <span class="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border" :class="form.mode === 'purge' ? 'border-rose-500' : 'border-gray-300'">
            <span v-if="form.mode === 'purge'" class="h-2 w-2 rounded-full bg-rose-500"></span>
          </span>
          <span>
            <span class="block text-sm font-medium text-gray-900">删除并清理图片</span>
            <span class="mt-0.5 block text-xs text-gray-500">永久删除密钥，并删除由该密钥上传的全部图片（文件与记录）。</span>
          </span>
        </button>
      </div>

      <!-- 警告 -->
      <div v-if="form.mode === 'revoke'" class="flex items-start gap-2 rounded-lg bg-amber-50 p-3 text-sm text-amber-700">
        <svg class="mt-0.5 h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <p>吊销后使用该密钥的客户端将立即无法访问。</p>
      </div>
      <div v-else class="flex items-start gap-2 rounded-lg bg-rose-50 p-3 text-sm text-rose-700">
        <svg class="mt-0.5 h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v3.75m-9-.75a9 9 0 1118 0 9 9 0 01-18 0zm9 3.75h.008v.008H12v-.008z" />
        </svg>
        <p>此操作不可撤销：密钥与关联图片将被永久删除。</p>
      </div>

      <!-- 账号密码二次确认 -->
      <div class="space-y-3 border-t border-gray-100 pt-4">
        <p class="text-sm font-medium text-gray-700">为确认是你本人操作，请重新输入账号密码</p>
        <div>
          <label for="rd-username" class="sr-only">用户名</label>
          <input
            id="rd-username"
            v-model="form.username"
            type="text"
            placeholder="用户名"
            autocomplete="username"
            class="w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-iris-dark focus:bg-white focus:ring-2 focus:ring-iris-dark/10"
          />
        </div>
        <div>
          <label for="rd-password" class="sr-only">密码</label>
          <input
            id="rd-password"
            v-model="form.password"
            type="password"
            placeholder="密码"
            autocomplete="current-password"
            class="w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-iris-dark focus:bg-white focus:ring-2 focus:ring-iris-dark/10"
          />
        </div>
        <div v-if="serverError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600">{{ serverError }}</div>
      </div>
    </div>

    <template #footer>
      <button type="button" class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50" @click="emit('close')">取消</button>
      <button
        type="button"
        :disabled="!canSubmit || loading"
        class="inline-flex items-center justify-center gap-1.5 rounded-xl px-3 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-60"
        :class="form.mode === 'purge' ? 'bg-rose-50 text-rose-600 hover:bg-rose-100' : 'bg-iris-violet/10 text-iris-dark hover:bg-iris-violet/15'"
        @click="handleSubmit"
      >
        <span v-if="loading" class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"></span>
        {{ form.mode === 'purge' ? '确认删除' : '确认吊销' }}
      </button>
    </template>
  </UiBaseDialog>
</template>

<script setup lang="ts">
import type { APIKeyInfo } from '~/composables/useApiKeys'

const props = defineProps<{ open: boolean; apiKey: APIKeyInfo | null }>()
const emit = defineEmits<{
  close: []
  done: [mode: 'revoke' | 'purge']
}>()

const { revoke, purge } = useApiKeys()
const { user } = useAuth()

type Mode = 'revoke' | 'purge'
const form = reactive({ mode: 'revoke' as Mode, username: '', password: '' })
const serverError = ref('')
const loading = ref(false)

const canSubmit = computed(() => form.username.trim() !== '' && form.password !== '')

watch(
  () => props.open,
  (v) => {
    if (v) {
      form.mode = 'revoke'
      // 预填当前登录用户名（仍可修改）。
      form.username = user.value?.username ?? ''
      form.password = ''
      serverError.value = ''
      loading.value = false
    }
  },
)

async function handleSubmit() {
  if (!props.apiKey || loading.value || !canSubmit.value) return
  loading.value = true
  serverError.value = ''
  try {
    const creds = { username: form.username.trim(), password: form.password }
    if (form.mode === 'purge') {
      await purge(props.apiKey.id, creds)
    } else {
      await revoke(props.apiKey.id, creds)
    }
    emit('done', form.mode)
  } catch (err: unknown) {
    serverError.value = err instanceof Error ? err.message : '操作失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
