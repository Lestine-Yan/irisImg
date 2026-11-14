<template>
  <div class="relative min-h-screen overflow-hidden antialiased">
    <LoginGeometricBackground />

    <div class="relative z-10 flex min-h-screen flex-col">
      <header
        class="mx-auto flex w-full max-w-7xl items-center justify-between px-6 pt-6 sm:px-10 lg:px-12 lg:pt-8"
      >
        <div class="flex items-center gap-2.5">
          <svg
            class="h-[18px] w-[18px] text-iris-dark"
            viewBox="0 0 24 24"
            fill="none"
            aria-hidden="true"
          >
            <path
              d="M12 21c0-5.5 3.2-9.8 8-11.5C19 15.2 15.8 19 12 21Z"
              fill="currentColor"
              opacity="0.85"
            />
            <path
              d="M12 21C12 15.5 8.8 11.2 4 9.5 5 15.2 8.2 19 12 21Z"
              fill="currentColor"
              opacity="0.35"
            />
          </svg>
          <span class="text-[13px] font-semibold tracking-[-0.01em] text-[#1C1A27]">irisImg</span>
          <span class="mx-0.5 hidden h-3 w-px bg-[#1C1A27]/15 sm:block" aria-hidden="true" />
          <span class="hidden text-[10px] uppercase tracking-[0.18em] text-[#8A8696] sm:inline">
            private cloud
          </span>
        </div>

        <div
          class="hidden items-center gap-2 rounded-full border border-white/60 bg-white/45 px-3 py-1.5 text-[9.5px] uppercase tracking-[0.24em] text-[#77737F] backdrop-blur-md sm:flex"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-500" />
          self-hosted
        </div>
      </header>

      <main class="flex flex-1 items-center">
        <div
          class="mx-auto grid w-full max-w-7xl gap-10 px-6 py-10 sm:px-10 lg:grid-cols-[minmax(0,1.08fr)_minmax(0,0.92fr)] lg:items-center lg:gap-10 lg:px-12 lg:py-8"
        >
          <LoginHero />

          <div class="flex justify-center lg:justify-end">
            <LoginForm @success="onLoginSuccess" />
          </div>
        </div>
      </main>

      <footer
        class="mx-auto flex w-full max-w-7xl items-center justify-between px-6 pb-6 text-[9.5px] uppercase tracking-[0.2em] text-[#A5A1AF] sm:px-10 lg:px-12"
      >
        <span>© {{ year }} irisImg</span>
        <span class="hidden sm:inline">为个人而建，由你所有</span>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: false,
})

const year = new Date().getFullYear()

function onLoginSuccess() {
  navigateTo('/dashboard')
}
</script>
