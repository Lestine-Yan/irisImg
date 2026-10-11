<template>
  <div />
</template>

<script setup lang="ts">
// 内容中心已拆分为 /content/img（图片）与 /content/audio（音频）两个子路由。
// 本页仅为兼容旧书签 / 旧入口的「即达图片页」重定向桩：
// 注意必须放在 pages/content/index.vue 而非 pages/content.vue——后者会注册成
// img/audio 的嵌套父路由组件，缺失 <NuxtPage> 时子页面将渲染空白。
definePageMeta({ middleware: 'auth' })

onMounted(() => {
  navigateTo('/content/img', { replace: true })
})
</script>
