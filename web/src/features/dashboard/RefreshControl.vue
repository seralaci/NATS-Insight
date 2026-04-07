<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  interval: number
}>()

const emit = defineEmits<{
  change: [seconds: number]
  refresh: []
}>()

const isOpen = ref(false)

const options = [
  { label: 'Manual', value: 0 },
  { label: '1s', value: 1 },
  { label: '2s', value: 2 },
  { label: '3s', value: 3 },
  { label: '5s', value: 5 },
  { label: '10s', value: 10 },
  { label: '30s', value: 30 },
  { label: '60s', value: 60 },
]

function select(value: number) {
  emit('change', value)
  isOpen.value = false
}

const currentLabel = () => options.find(o => o.value === props.interval)?.label ?? 'Manual'
</script>

<template>
  <div class="flex items-center gap-2">
    <!-- Refresh button -->
    <button
      class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400"
      title="Refresh now"
      @click="$emit('refresh')"
    >
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182"/>
      </svg>
    </button>

    <!-- Refresh mode dropdown -->
    <div class="relative">
      <button
        class="flex items-center gap-1 px-2 py-1 text-xs text-gray-500 dark:text-gray-400 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
        @click="isOpen = !isOpen"
      >
        <span>Refresh Mode:</span>
        <span class="font-medium text-gray-700 dark:text-gray-300">{{ currentLabel() }}</span>
        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5"/>
        </svg>
      </button>

      <div
        v-if="isOpen"
        class="absolute right-0 top-full mt-1 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-md shadow-lg z-10 py-1 min-w-24"
      >
        <button
          v-for="opt in options"
          :key="opt.value"
          class="w-full text-left px-3 py-1 text-xs hover:bg-gray-100 dark:hover:bg-gray-800"
          :class="opt.value === interval ? 'text-emerald-600 dark:text-emerald-400 font-medium' : 'text-gray-700 dark:text-gray-300'"
          @click="select(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
    </div>
  </div>
</template>
