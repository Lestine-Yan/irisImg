<template>
  <div class="flex items-start justify-between gap-4 py-3">
    <dt class="text-sm text-gray-500">{{ label }}</dt>
    <dd class="text-right text-sm font-medium text-gray-900">
      <slot />
    </dd>
  </div>
</template>

<script setup lang="ts">
defineProps<{ label: string }>()
</script>
