<script setup lang="ts">
import { computed } from 'vue'
import { useObjectsStore } from '../../stores/objects'

const objectsStore = useObjectsStore()

const store = computed(() => objectsStore.selectedStore)

function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return 'Unlimited'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}
</script>

<template>
  <div class="flex-1 overflow-auto p-6">
    <div v-if="!store" class="text-sm text-gray-400 dark:text-gray-600 italic">No store selected</div>

    <div v-else class="grid grid-cols-3 gap-6">
      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">General</h3>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Store Name</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ store.bucket }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Description</span>
            <span class="text-gray-700 dark:text-gray-300">{{ store.description || '-' }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Storage</span>
            <span class="font-mono text-gray-800 dark:text-gray-200 capitalize">{{ store.storage }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Replicas</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ store.replicas }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Sealed</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ store.sealed ? 'Yes' : 'No' }}</span>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Limits</h3>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Chunk Size</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">-</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Bytes</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatBytes(store.size) }}</span>
          </div>
        </div>
      </div>

      <div
        v-if="store.metadata && Object.keys(store.metadata).length > 0"
        class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4"
      >
        <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Metadata</h3>
        <div class="space-y-2">
          <div v-for="(val, key) in store.metadata" :key="key" class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">{{ key }}</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ val }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
