<template>
  <div class="login-card relative w-full max-w-[420px]">
    <div
      class="absolute -inset-5 -z-10 rounded-[36px] bg-gradient-to-br from-white/70 via-white/20 to-iris-violet/10 blur-2xl"
      aria-hidden="true"
    />

    <div
      class="relative overflow-hidden rounded-[26px] border border-white/90 bg-white/90 p-6 shadow-[0_28px_80px_-36px_rgba(42,33,80,0.38),0_2px_8px_rgba(42,33,80,0.04)] backdrop-blur-xl sm:p-9"
    >
      <div
        class="pointer-events-none absolute inset-x-10 top-0 h-px bg-gradient-to-r from-transparent via-white to-transparent"
        aria-hidden="true"
      />

      <div class="flex items-center justify-between gap-4">
        <p class="text-[10px] font-semibold uppercase tracking-[0.3em] text-[#8A8696]">
          Sign in
        </p>
        <span
          class="inline-flex items-center gap-1.5 rounded-full border border-iris-dark/10 bg-iris-dark/[0.04] px-2.5 py-1 text-[10px] font-medium tracking-wide text-iris-dark/80"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-500" />
          私有部署
        </span>
      </div>

      <h2 class="mt-5 text-[26px] font-semibold tracking-[-0.02em] text-[#1C1A27]">
        欢迎回来
      </h2>
      <p class="mt-2 text-[13px] leading-6 text-[#77737F]">
        登录以进入你的私人存储空间。
      </p>

      <form class="mt-8 space-y-5" novalidate @submit.prevent="handleSubmit">
        <div>
          <div class="mb-2 flex items-baseline justify-between">
            <label for="username" class="text-[12px] font-medium text-[#514D5C]">用户名</label>
            <span class="text-[9px] uppercase tracking-[0.18em] text-[#ADA9B6]">Username</span>
          </div>
          <div class="relative">
            <svg
              class="pointer-events-none absolute left-4 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-[#A5A1AF]"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
              />
            </svg>
            <input
              id="username"
              v-model="form.username"
              type="text"
              autocomplete="username"
              autocapitalize="off"
              spellcheck="false"
              placeholder="请输入用户名"
              :aria-invalid="Boolean(errors.username)"
              :aria-describedby="errors.username ? 'username-error' : undefined"
              class="h-12 w-full rounded-xl border border-[#E4E2EA] bg-white/60 pl-11 pr-4 text-[14px] text-[#1C1A27] outline-none transition duration-200 placeholder:text-[#ADA9B6] hover:border-[#D6D3DE] focus:border-iris-dark/45 focus:bg-white focus:ring-4 focus:ring-iris-dark/[0.06]"
              :class="{ 'border-red-300 focus:border-red-400 focus:ring-red-500/10': errors.username }"
              @input="clearError('username')"
            />
          </div>
          <p
            v-if="errors.username"
            id="username-error"
            class="mt-1.5 flex items-center gap-1 text-[11px] text-red-500"
          >
            <svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v4m0 4h.01M10.3 3.9L1.8 18a2 2 0 001.7 3h17a2 2 0 001.7-3L13.7 3.9a2 2 0 00-3.4 0z" />
            </svg>
            {{ errors.username }}
          </p>
        </div>

        <div>
          <div class="mb-2 flex items-baseline justify-between">
            <label for="password" class="text-[12px] font-medium text-[#514D5C]">密码</label>
            <span class="text-[9px] uppercase tracking-[0.18em] text-[#ADA9B6]">Password</span>
          </div>
          <div class="relative">
            <svg
              class="pointer-events-none absolute left-4 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-[#A5A1AF]"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75c0 1.24 1.01 2.25 2.25 2.25z"
              />
            </svg>
            <input
              id="password"
              v-model="form.password"
              :type="passwordType"
              autocomplete="current-password"
              placeholder="请输入密码"
              :aria-invalid="Boolean(errors.password)"
              :aria-describedby="errors.password ? 'password-error' : undefined"
              class="h-12 w-full rounded-xl border border-[#E4E2EA] bg-white/60 pl-11 pr-12 text-[14px] text-[#1C1A27] outline-none transition duration-200 placeholder:text-[#ADA9B6] hover:border-[#D6D3DE] focus:border-iris-dark/45 focus:bg-white focus:ring-4 focus:ring-iris-dark/[0.06]"
              :class="{ 'border-red-300 focus:border-red-400 focus:ring-red-500/10': errors.password }"
              @input="clearError('password')"
            />
            <button
              type="button"
              class="absolute right-2 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-lg text-[#A5A1AF] transition hover:bg-[#1C1A27]/[0.04] hover:text-[#514D5C] focus:outline-none focus-visible:ring-2 focus-visible:ring-iris-dark/30"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
              :aria-pressed="showPassword"
              @click="showPassword = !showPassword"
            >
              <svg v-if="!showPassword" class="h-[18px] w-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.04 12.32a1.01 1.01 0 010-.64C3.42 7.51 7.36 4.5 12 4.5s8.58 3 9.96 7.18c.07.21.07.43 0 .64C20.58 16.49 16.64 19.5 12 19.5s-8.58-3-9.96-7.18z" />
                <circle cx="12" cy="12" r="3" stroke-width="1.5" />
              </svg>
              <svg v-else class="h-[18px] w-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 3l18 18M10.6 10.7a2 2 0 002.8 2.8M9.9 5.1A10.6 10.6 0 0112 4.9c4.64 0 8.58 3 9.96 7.18.07.21.07.43 0 .64a11.6 11.6 0 01-2.4 3.9M6.2 6.4a11.8 11.8 0 00-4.16 5.28 1.01 1.01 0 000 .64C3.42 16.49 7.36 19.5 12 19.5c1.05 0 2.06-.15 3-.43" />
              </svg>
            </button>
          </div>
          <p
            v-if="errors.password"
            id="password-error"
            class="mt-1.5 flex items-center gap-1 text-[11px] text-red-500"
          >
            <svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v4m0 4h.01M10.3 3.9L1.8 18a2 2 0 001.7 3h17a2 2 0 001.7-3L13.7 3.9a2 2 0 00-3.4 0z" />
            </svg>
            {{ errors.password }}
          </p>
        </div>

        <div
          v-if="serverError"
          role="alert"
          class="flex items-start gap-2 rounded-xl border border-red-200/70 bg-red-50/80 px-3.5 py-3 text-[12px] leading-5 text-red-600"
        >
          <svg class="mt-0.5 h-3.5 w-3.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v4m0 4h.01M10.3 3.9L1.8 18a2 2 0 001.7 3h17a2 2 0 001.7-3L13.7 3.9a2 2 0 00-3.4 0z" />
          </svg>
          <span>{{ serverError }}</span>
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="group flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-[#211B33] px-4 text-[14px] font-semibold tracking-wide text-white shadow-[0_12px_30px_-14px_rgba(109,79,216,0.7)] transition duration-200 hover:bg-iris-dark hover:shadow-[0_16px_36px_-14px_rgba(109,79,216,0.8)] focus:outline-none focus-visible:ring-2 focus-visible:ring-iris-dark/40 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <span v-if="loading" class="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
          <span>{{ loading ? '登录中...' : '进入工作台' }}</span>
          <svg
            v-if="!loading"
            class="h-4 w-4 transition-transform duration-200 group-hover:translate-x-0.5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M4 12h15m0 0l-5.5-5.5M19 12l-5.5 5.5" />
          </svg>
        </button>
      </form>

      <div class="mt-6 flex items-center justify-between gap-3 border-t border-[#1C1A27]/[0.07] pt-4">
        <span class="flex items-center gap-1.5 text-[10.5px] tracking-wide text-[#8A8696]">
          <svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.6" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75c0 1.24 1.01 2.25 2.25 2.25z" />
          </svg>
          凭据仅提交至你的服务器
        </span>
        <span class="hidden items-center gap-1 text-[10px] text-[#ADA9B6] sm:flex">
          <kbd class="rounded border border-[#1C1A27]/10 bg-white/70 px-1.5 py-0.5 font-sans text-[9px] text-[#8A8696]">Enter</kbd>
          登录
        </span>
      </div>
    </div>
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
const showPassword = ref(false)

const passwordType = computed(() => (showPassword.value ? 'text' : 'password'))

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

<style scoped>
.login-card {
  animation: card-in 0.8s cubic-bezier(0.22, 1, 0.36, 1) 0.08s both;
}

@keyframes card-in {
  from {
    opacity: 0;
    transform: translateY(18px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .login-card {
    animation: none;
  }
}
</style>
