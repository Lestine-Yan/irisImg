<template>
  <div class="hero-reveal flex max-w-xl flex-col items-start">
    <div class="flex items-center gap-3">
      <span class="h-px w-8 bg-iris-dark/30" />
      <p class="text-[10px] font-semibold uppercase tracking-[0.32em] text-iris-dark/65">
        Private object storage
      </p>
    </div>

    <h1 class="mt-5 sm:mt-7">
      <span class="sr-only">irisImg —— 轻量、高性能的私人对象存储</span>
      <svg
        aria-hidden="true"
        viewBox="0 0 250 200"
        class="block h-auto w-[158px] sm:w-[232px] lg:w-[278px] xl:w-[306px]"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <linearGradient id="irisWordmarkGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#6D4FD8" />
            <stop offset="58%" stop-color="#8F7BE8" />
            <stop offset="100%" stop-color="#86C9EC" />
          </linearGradient>
        </defs>
        <text
          x="2"
          y="84"
          font-family="Inter, system-ui, -apple-system, BlinkMacSystemFont, sans-serif"
          font-size="96"
          font-weight="800"
          letter-spacing="-4"
          fill="#1C1A27"
        >iris</text>
        <text
          x="2"
          y="176"
          font-family="Inter, system-ui, -apple-system, BlinkMacSystemFont, sans-serif"
          font-size="96"
          font-weight="800"
          letter-spacing="-4"
          fill="url(#irisWordmarkGradient)"
        >Img</text>
      </svg>
    </h1>

    <p class="mt-6 text-[22px] font-semibold leading-[1.35] tracking-[-0.01em] text-[#1C1A27] sm:mt-8 sm:text-[28px] xl:text-[32px]">
      轻量、高性能的
      <br class="hidden sm:block" />
      <span class="text-iris-dark">私人对象存储</span>
    </p>

    <p class="mt-3 max-w-md text-[14px] leading-[1.9] text-[#6B6775] sm:mt-4 sm:text-[15px]">
      为图片与音频打造的自托管空间。
      <br />
      上传、整理、直链分发，数据始终由你掌控。
    </p>

    <dl class="mt-10 hidden w-full max-w-lg gap-x-6 border-t border-[#1C1A27]/[0.08] pt-6 sm:grid sm:grid-cols-3">
      <div v-for="item in features" :key="item.label">
        <dt class="flex items-center gap-2 text-[12.5px] font-semibold text-[#2A2733]">
          <span class="h-1 w-1 rounded-full bg-iris-dark" />
          {{ item.label }}
        </dt>
        <dd class="mt-1.5 text-[11.5px] leading-5 text-[#807C8A]">
          {{ item.detail }}
        </dd>
      </div>
    </dl>
  </div>
</template>

<script setup lang="ts">
const features = [
  { label: '轻量', detail: 'Go 单二进制，解压即用' },
  { label: '高性能', detail: '静态直链，Nginx 直接分发' },
  { label: '私有', detail: '单用户模型，数据自持' },
]
</script>

<style scoped>
.hero-reveal {
  animation: hero-in 0.8s cubic-bezier(0.22, 1, 0.36, 1) both;
}

@keyframes hero-in {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .hero-reveal {
    animation: none;
  }
}
</style>
