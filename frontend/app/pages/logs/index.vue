<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">日志中心</h1>
      <p class="mt-1 text-sm text-gray-500">查看图片上传与访问日志</p>
    </div>

    <div class="rounded-2xl border border-dashed border-gray-300 bg-white/60 p-12 text-center">
      <p class="text-sm text-gray-400">该模块正在开发中</p>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ middleware: 'auth' })
</script>
