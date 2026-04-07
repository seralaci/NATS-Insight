<script setup lang="ts">
import { computed } from 'vue'
import { useKvStore } from '../../stores/kv'

const kvStore = useKvStore()

const bucket = computed(() => kvStore.selectedBucket)

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

function formatTtl(ns: number): string {
  if (!ns) return '-'
  const ms = ns / 1_000_000
  const secs = ms / 1000
  if (secs < 60) return `${secs}s`
  if (secs < 3600) return `${Math.round(secs / 60)}m`
  if (secs < 86400) return `${Math.round(secs / 3600)}h`
  return `${Math.round(secs / 86400)}d`
}
</script>

<template>
  <div class="flex-1 overflow-auto p-6">
    <div v-if="!bucket" class="text-sm text-gray-400 dark:text-gray-600 italic">No bucket selected</div>

    <div v-else class="grid grid-cols-3 gap-6">
      <!-- General -->
      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">General</h3>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Bucket Name</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ bucket.name }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Description</span>
            <span class="text-gray-700 dark:text-gray-300">{{ bucket.description || '-' }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">History Values</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ bucket.history }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Storage</span>
            <span class="font-mono text-gray-800 dark:text-gray-200 capitalize">{{ bucket.storage }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Compression</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ bucket.isCompressed ? 'S2' : 'None' }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Replicas</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ bucket.replicas }}</span>
          </div>
        </div>
      </div>

      <!-- State -->
      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">State</h3>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Values Stored</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ bucket.values }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Bucket Size</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatBytes(bucket.bytes) }}</span>
          </div>
        </div>
      </div>

      <!-- Limits -->
      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Limits</h3>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Bucket Size</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">Unlimited</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Value Size</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">Unlimited</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Maximum Age</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatTtl(bucket.ttl) }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Limit Marker TTL</span>
            <span class="font-mono text-gray-800 dark:text-gray-200">-</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
