<script setup lang="ts">
import { computed } from 'vue'
import { useStreamsStore } from '../../stores/streams'

const streamsStore = useStreamsStore()

const stream = computed(() => streamsStore.selectedStream)

function formatBytes(bytes: number): string {
  if (!bytes || bytes < 0) return 'Unlimited'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

function formatAge(ns: number): string {
  if (!ns || ns <= 0) return 'Unlimited'
  const secs = ns / 1e9
  if (secs < 60) return `${secs}s`
  if (secs < 3600) return `${Math.round(secs / 60)}m`
  if (secs < 86400) return `${Math.round(secs / 3600)}h`
  return `${Math.round(secs / 86400)}d`
}

function formatDuplicates(ns: number): string {
  if (!ns || ns <= 0) return '-'
  const secs = ns / 1e9
  if (secs < 60) return `${secs}s`
  if (secs < 3600) return `${Math.round(secs / 60)}m`
  return `${Math.round(secs / 3600)}h`
}

function formatDate(iso: string): string {
  if (!iso) return '-'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function unlimited(val: number): string {
  if (val === undefined || val === null || val < 0) return 'Unlimited'
  return String(val)
}
</script>

<template>
  <div class="flex-1 overflow-auto p-6">
    <div v-if="!stream" class="text-sm text-gray-400 dark:text-gray-600 italic">No stream selected</div>

    <div v-else class="space-y-6">
      <!-- Row 1: General + State + Replication -->
      <div class="grid grid-cols-3 gap-6">
        <!-- General -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">General</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Name</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.name }}</span>
            </div>
            <div v-if="stream.config.subjects?.length" class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Subjects</span>
              <span class="font-mono text-gray-800 dark:text-gray-200 text-right max-w-xs">{{ stream.config.subjects.join(', ') }}</span>
            </div>
            <div v-if="stream.config.description" class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Description</span>
              <span class="text-gray-700 dark:text-gray-300">{{ stream.config.description }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Retention Policy</span>
              <span class="font-mono text-gray-800 dark:text-gray-200 capitalize">{{ stream.config.retention }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Storage</span>
              <span class="font-mono text-gray-800 dark:text-gray-200 capitalize">{{ stream.config.storage }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Compression</span>
              <span class="font-mono text-gray-800 dark:text-gray-200 uppercase">{{ stream.config.compression || 'None' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Replicas</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.replicas }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Sealed</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.sealed ? 'true' : 'false' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Created</span>
              <span class="text-gray-700 dark:text-gray-300 text-xs">{{ formatDate(stream.created) }}</span>
            </div>
          </div>
        </div>

        <!-- State -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">State</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Messages</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.state.messages }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Consumers</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.state.consumer_count }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Size</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatBytes(stream.state.bytes) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">First Sequence</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.state.first_seq || '---' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Last Sequence</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.state.last_seq || '---' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Lost Messages</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.state.num_deleted ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Num Subjects</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.state.num_subjects ?? 0 }}</span>
            </div>
          </div>
        </div>

        <!-- Metadata -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Metadata</h3>
          <div v-if="stream.config.metadata && Object.keys(stream.config.metadata).length > 0" class="space-y-2">
            <div v-for="(val, key) in stream.config.metadata" :key="key" class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide font-mono">{{ key }}</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val }}</span>
            </div>
          </div>
          <div v-else class="text-xs text-gray-400 dark:text-gray-600 italic">No metadata</div>
        </div>
      </div>

      <!-- Row 2: Options + Limits -->
      <div class="grid grid-cols-2 gap-6">
        <!-- Options -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Options</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Discard Policy</span>
              <span class="font-mono text-gray-800 dark:text-gray-200 capitalize">{{ stream.config.discard }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Discard New Per Subject</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.discardNewPerSubject ? 'true' : 'false' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Duplicate Window</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatDuplicates(stream.config.duplicates) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">No Acknowledgement</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.noAck ? 'true' : 'false' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Allow Rollup Headers</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.allowRollup ? 'true' : 'false' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Allow Direct Access</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.allowDirect ? 'true' : 'false' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Deny Delete</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.denyDelete ? 'true' : 'false' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Deny Purge</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.denyPurge ? 'true' : 'false' }}</span>
            </div>
          </div>
        </div>

        <!-- Limits -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Limits</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Messages</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ unlimited(stream.config.maxMsgs) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Messages Per Subject</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ unlimited(stream.config.maxMsgsPerSubject) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Bytes</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.maxBytes && stream.config.maxBytes > 0 ? formatBytes(stream.config.maxBytes) : 'Unlimited' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Message Size</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ stream.config.maxMsgSize && stream.config.maxMsgSize > 0 ? formatBytes(stream.config.maxMsgSize) : 'Unlimited' }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Age</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatAge(stream.config.maxAge) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-500 dark:text-gray-400 uppercase text-xs tracking-wide">Max Consumers</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ unlimited(stream.config.maxConsumers) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
