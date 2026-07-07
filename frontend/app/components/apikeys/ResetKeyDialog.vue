<template>
  <UiBaseDialog :open="open" title="重置 API Key 明文" @close="emit('close')">
    <div class="space-y-4">
      <p class="text-sm text-gray-600">
        重置后会生成新的明文密钥并替换旧密钥，旧明文将立即失效。此操作也会取消密钥的吊销状态（重新激活）。
      </p>
      <div class="flex items-start gap-2 rounded-lg bg-amber-50 p-3 text-sm text-amber-700">
        <svg class="mt-0.5 h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <p>重置后的新明文仅显示一次，请提前准备好保存。</p>
      </div>
      <div v-if="serverError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600">{{ serverError }}</div>
    </div>

    <template #footer>
      <button type="button" class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50" @click="emit('close')">取消</button>
      <button
        type="button"
        :disabled="loading"
        class="inline-flex items-center justify-center gap-1.5 rounded-xl bg-iris-violet/10 px-3 py-2 text-sm font-medium text-iris-dark transition hover:bg-iris-violet/15 disabled:cursor-not-allowed disabled:opacity-60"
        @click="handleReset"
      >
        <span v-if="loading" class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"></span>
        确认重置
      </button>
    </template>
  </UiBaseDialog>
</template>

<script setup lang="ts">
import type { APIKeyInfo, ResetAPIKeyResponse } from '~/composables/useApiKeys'

const props = defineProps<{ open: boolean; apiKey: APIKeyInfo | null }>()
const emit = defineEmits<{
  close: []
  reset: [resp: ResetAPIKeyResponse]
}>()

const { reset } = useApiKeys()

const serverError = ref('')
const loading = ref(false)

watch(
  () => props.open,
  (v) => {
    if (v) {
      serverError.value = ''
      loading.value = false
    }
  },
)

async function handleReset() {
  if (!props.apiKey || loading.value) return
  loading.value = true
  serverError.value = ''
  try {
    const resp = await reset(props.apiKey.id)
    emit('reset', resp)
  } catch (err: unknown) {
    serverError.value = err instanceof Error ? err.message : '重置失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
