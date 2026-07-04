<template>
  <div class="w-full max-w-md rounded-2xl bg-white p-8 shadow-2xl md:p-10">
    <h2 class="mb-8 text-2xl font-bold text-gray-900">欢迎回来</h2>

    <form class="space-y-5" @submit.prevent="handleSubmit">
      <div>
        <label for="username" class="sr-only">用户名</label>
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
            <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <input
            id="username"
            v-model="form.username"
            type="text"
            placeholder="请输入用户名"
            class="w-full rounded-xl border border-gray-200 bg-gray-50 py-3 pl-12 pr-4 text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-iris-dark focus:bg-white focus:ring-2 focus:ring-iris-dark/10"
            :class="{ 'border-red-300 focus:border-red-500 focus:ring-red-500/10': errors.username }"
            @input="clearError('username')"
          />
        </div>
        <p v-if="errors.username" class="mt-1.5 text-sm text-red-500">{{ errors.username }}</p>
      </div>

      <div>
        <label for="password" class="sr-only">密码</label>
        <div class="relative">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
            <svg class="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75c0 1.24 1.01 2.25 2.25 2.25z" />
            </svg>
          </div>
          <input
            id="password"
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            class="w-full rounded-xl border border-gray-200 bg-gray-50 py-3 pl-12 pr-4 text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-iris-dark focus:bg-white focus:ring-2 focus:ring-iris-dark/10"
            :class="{ 'border-red-300 focus:border-red-500 focus:ring-red-500/10': errors.password }"
            @input="clearError('password')"
          />
        </div>
        <p v-if="errors.password" class="mt-1.5 text-sm text-red-500">{{ errors.password }}</p>
      </div>

      <div v-if="serverError" class="rounded-lg bg-red-50 p-3 text-sm text-red-600">
        {{ serverError }}
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="flex w-full items-center justify-center rounded-xl bg-iris-dark py-3.5 text-base font-semibold text-white shadow-lg transition-all hover:bg-iris-violet hover:shadow-xl focus:outline-none focus:ring-2 focus:ring-iris-dark/20 disabled:cursor-not-allowed disabled:opacity-60"
      >
        <span v-if="loading" class="mr-2 h-5 w-5 animate-spin rounded-full border-2 border-white/30 border-t-white"></span>
        {{ loading ? '登录中...' : '进入工作台' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
interface FormState {
  username: string
  password: string
}

interface FormErrors {
  username: string
  password: string
}

const emit = defineEmits<{
  success: []
}>()

const { login } = useAuth()

const form = reactive<FormState>({
  username: '',
  password: '',
})

const errors = reactive<FormErrors>({
  username: '',
  password: '',
})

const serverError = ref('')
const loading = ref(false)

function clearError(field: keyof FormErrors) {
  errors[field] = ''
  serverError.value = ''
}

function validate(): boolean {
  errors.username = ''
  errors.password = ''
  let valid = true

  if (!form.username.trim()) {
    errors.username = '请输入用户名'
    valid = false
  }

  if (!form.password) {
    errors.password = '请输入密码'
    valid = false
  } else if (form.password.length < 4) {
    errors.password = '密码长度不能少于 4 位'
    valid = false
  }

  return valid
}

async function handleSubmit() {
  if (!validate()) return

  loading.value = true
  serverError.value = ''

  try {
    await login(form.username.trim(), form.password)
    emit('success')
  } catch (err: unknown) {
    if (err instanceof Error) {
      serverError.value = err.message
    } else {
      serverError.value = '登录失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}
</script>
