<script setup lang="ts">
defineProps<{
  label: string
  value: string | number
}>()
</script>

<template>
  <div class="bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
    <div class="text-2xl font-bold text-gray-900 dark:text-gray-100">{{ value }}</div>
    <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ label }}</div>
  </div>
</template>
