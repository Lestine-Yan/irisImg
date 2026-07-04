<template>
  <div class="relative min-h-screen">
    <LoginGeometricBackground />

    <div class="relative z-10 mx-auto flex min-h-screen max-w-7xl flex-col items-center justify-center px-6 py-12 lg:flex-row lg:justify-between">
      <div class="mb-12 w-full lg:mb-0 lg:w-1/2 lg:self-start lg:pr-12 lg:pt-24">
        <LoginHero />
      </div>

      <div class="flex w-full justify-center lg:w-1/2 lg:justify-end">
        <LoginForm @success="onLoginSuccess" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: false,
})

function onLoginSuccess() {
  navigateTo('/dashboard')
}
</script>
