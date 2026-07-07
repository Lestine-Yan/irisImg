<template>
  <UiBaseDialog :open="open" title="创建 API Key" @close="emit('close')">
    <form class="space-y-4" @submit.prevent="handleSubmit">
      <div>
        <label for="api-key-name" class="block text-sm font-medium text-gray-700">名称</label>
        <input
          id="api-key-name"
          v-model="form.name"
          type="text"
          placeholder="如：博客图床上传"
          class="mt-1 w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-iris-dark focus:bg-white focus:ring-2 focus:ring-iris-dark/10"
          :class="{ 'border-red-300 focus:border-red-500 focus:ring-red-500/10': errors.name }"
          @input="errors.name = ''"
        />
        <p v-if="errors.name" class="mt-1.5 text-sm text-red-500">{{ errors.name }}</p>
      </div>

      <div>
        <span class="block text-sm font-medium text-gray-700">权限范围</span>
        <div class="mt-2 grid grid-cols-2 gap-2">
          <button
            v-for="opt in scopeOptions"
            :key="opt.value"
            type="button"
            class="rounded-xl border px-3 py-2.5 text-left text-sm transition"
            :class="
              form.scope === opt.value
                ? 'border-iris-dark bg-iris-violet/10 text-iris-dark'
                : 'border-gray-200 text-gray-600 hover:bg-gray-50'
            "
            @click="form.scope = opt.value"
          >
            <p class="font-medium">{{ opt.label }}</p>
            <p class="mt-0.5 text-xs text-gray-400">{{ opt.desc }}</p>
          </button>
        </div>
      </div>

      <div v-if="serverError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600">{{ serverError }}</div>
    </form>

    <template #footer>
      <button
        type="button"
        class="rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50"
        @click="emit('close')"
      >
        取消
      </button>
      <button
        type="button"
        :disabled="loading"
        class="inline-flex items-center justify-center gap-1.5 rounded-xl bg-iris-violet/10 px-3 py-2 text-sm font-medium text-iris-dark transition hover:bg-iris-violet/15 disabled:cursor-not-allowed disabled:opacity-60"
        @click="handleSubmit"
      >
        <span v-if="loading" class="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white"></span>
        创建
      </button>
    </template>
  </UiBaseDialog>
</template>

<script setup lang="ts">
import type { APIKeyScope, CreateAPIKeyResponse } from '~/composables/useApiKeys'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{
  close: []
  created: [resp: CreateAPIKeyResponse]
}>()

const { create } = useApiKeys()

const scopeOptions = [
  { value: 'readwrite' as APIKeyScope, label: '读写', desc: '可上传与获取图片' },
  { value: 'readonly' as APIKeyScope, label: '只读', desc: '仅可获取图片' },
]

const form = reactive({ name: '', scope: 'readwrite' as APIKeyScope })
const errors = reactive({ name: '' })
const serverError = ref('')
const loading = ref(false)

// 每次打开时重置表单。
watch(
  () => props.open,
  (v) => {
    if (v) {
      form.name = ''
      form.scope = 'readwrite'
      errors.name = ''
      serverError.value = ''
      loading.value = false
    }
  },
)

function validate(): boolean {
  errors.name = ''
  if (!form.name.trim()) {
    errors.name = '请输入名称'
    return false
  }
  return true
}

async function handleSubmit() {
  if (!validate() || loading.value) return
  loading.value = true
  serverError.value = ''
  try {
    const resp = await create({ name: form.name.trim(), scope: form.scope })
    emit('created', resp)
  } catch (err: unknown) {
    serverError.value = err instanceof Error ? err.message : '创建失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
