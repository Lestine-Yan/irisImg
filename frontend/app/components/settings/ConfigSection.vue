<template>
  <section class="admin-surface p-6">
    <h2 class="text-sm font-semibold uppercase tracking-wide text-gray-500">{{ title }}</h2>
    <dl class="mt-4 divide-y divide-gray-100">
      <slot />
    </dl>
  </section>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>
