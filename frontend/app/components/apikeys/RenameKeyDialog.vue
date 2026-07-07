<template>
  <UiBaseDialog :open="open" title="重命名 API Key" @close="emit('close')">
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div>
        <label for="rename-name" class="block text-sm font-medium text-gray-700">新名称</label>
        <input
          id="rename-name"
          v-model="form.name"
          type="text"
          class="mt-1 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-gray-900 outline-none transition-colors focus:border-iris-dark focus:bg-white focus:ring-2 focus:ring-iris-dark/10"
          :class="{ 'border-red-300 focus:border-red-500 focus:ring-red-500/10': errors.name }"
          @input="errors.name = ''"
        />
        <p v-if="errors.name" class="mt-1.5 text-sm text-red-500">{{ errors.name }}</p>
      </div>
      <div v-if="serverError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600">{{ serverError }}</div>
    </form>

    <template #footer>
      <button type="button" class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50" @click="emit('close')">取消</button>
      <button
        type="button"
        :disabled="loading"
        class="inline-flex items-center justify-center gap-1.5 rounded-xl bg-iris-violet/10 px-3 py-2 text-sm font-medium text-iris-dark transition hover:bg-iris-violet/15 disabled:cursor-not-allowed disabled:opacity-60"
        @click="handleSubmit"
      >
        <span v-if="loading" class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"></span>
        保存
      </button>
    </template>
  </UiBaseDialog>
</template>

<script setup lang="ts">
import type { APIKeyInfo } from '~/composables/useApiKeys'

const props = defineProps<{ open: boolean; apiKey: APIKeyInfo | null }>()
const emit = defineEmits<{
  close: []
  renamed: []
}>()

const { rename } = useApiKeys()

const form = reactive({ name: '' })
const errors = reactive({ name: '' })
const serverError = ref('')
const loading = ref(false)

watch(
  () => props.open,
  (v) => {
    if (v) {
      form.name = props.apiKey?.name ?? ''
      errors.name = ''
      serverError.value = ''
      loading.value = false
    }
  },
)

async function handleSubmit() {
  if (!props.apiKey || loading.value) return
  if (!form.name.trim()) {
    errors.name = '请输入名称'
    return
  }
  loading.value = true
  serverError.value = ''
  try {
    await rename(props.apiKey.id, form.name.trim())
    emit('renamed')
  } catch (err: unknown) {
    serverError.value = err instanceof Error ? err.message : '重命名失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
