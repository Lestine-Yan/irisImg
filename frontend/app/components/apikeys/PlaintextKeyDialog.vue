<template>
  <UiBaseDialog :open="open" :title="title" @close="emit('close')">
    <div class="space-y-4">
      <div class="flex items-start gap-2 rounded-lg bg-amber-50 p-3 text-sm text-amber-700">
        <svg class="mt-0.5 h-5 w-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
        <p>该明文仅此一次显示，请立即复制并妥善保存；关闭后无法再次查看。</p>
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700">明文 API Key</label>
        <div class="mt-1 flex gap-2">
          <input
            ref="keyInput"
            :value="plaintext"
            readonly
            class="w-full rounded-xl border border-gray-200 bg-gray-50 px-4 py-2.5 font-mono text-sm text-gray-800 outline-none focus:border-iris-dark focus:ring-2 focus:ring-iris-dark/10"
            @focus="($event.target as HTMLInputElement)?.select()"
          />
          <button
            type="button"
            class="inline-flex shrink-0 items-center gap-1.5 rounded-xl px-3 py-2 text-sm font-medium transition"
            :class="copied ? 'bg-emerald-100 text-emerald-700' : 'bg-iris-violet/10 text-iris-dark hover:bg-iris-violet/15'"
            @click="copy"
          >
            <svg v-if="copied" class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4.5 12.75l6 6 9-13.5" />
            </svg>
            <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 01-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 011.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.582-8.25-8-8.25-.621 0-1.125.504-1.125 1.125v3.375m9.75 10.376L9.75 12.75m0 0L6 9m3.75 3.75v6" />
            </svg>
            {{ copied ? '已复制' : '复制' }}
          </button>
        </div>
      </div>
    </div>

    <template #footer>
      <button
        type="button"
        class="inline-flex items-center justify-center gap-1.5 rounded-xl bg-iris-violet/10 px-3 py-2 text-sm font-medium text-iris-dark transition hover:bg-iris-violet/15"
        @click="emit('close')"
      >
        我已保存
      </button>
    </template>
  </UiBaseDialog>
</template>

<script setup lang="ts">
const props = defineProps<{
  open: boolean
  plaintext: string
  title?: string
}>()
const emit = defineEmits<{ close: [] }>()

const copied = ref(false)
const keyInput = ref<HTMLInputElement | null>(null)

async function copy() {
  if (!props.plaintext) return
  try {
    await navigator.clipboard.writeText(props.plaintext)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch {
    // clipboard 不可用时兜底选中输入框，便于手动复制。
    keyInput.value?.focus()
    keyInput.value?.select()
  }
}

// 每次打开重置「已复制」状态。
watch(
  () => props.open,
  (v) => {
    if (v) copied.value = false
  },
)
</script>
