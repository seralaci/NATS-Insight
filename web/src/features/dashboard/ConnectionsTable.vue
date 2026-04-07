<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { serverApi } from '../../lib/api'

interface ConnzConnection {
  cid: number
  name: string
  ip: string
  port: number
  subs: number
  in_msgs: number
  out_msgs: number
  in_bytes: number
  out_bytes: number
  uptime: string
}

interface ConnzResponse {
  connections: ConnzConnection[]
  num_connections: number
}

const connections = ref<ConnzConnection[]>([])
const loading = ref(false)
const error = ref('')

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

async function fetchConnections() {
  loading.value = true
  error.value = ''
  try {
    const data = await serverApi.connz() as ConnzResponse
    connections.value = data.connections ?? []
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchConnections())

defineExpose({ refresh: fetchConnections })
</script>

<template>
  <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden">
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-800">
      <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Active Connections</h3>
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-40"
        :disabled="loading"
        title="Refresh"
        @click="fetchConnections"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
        </svg>
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-10">
      <span class="text-sm text-gray-400 dark:text-gray-600">Loading connections...</span>
    </div>

    <div v-else-if="error" class="px-4 py-4 text-sm text-red-600 dark:text-red-400">{{ error }}</div>

    <div v-else-if="connections.length === 0" class="flex items-center justify-center py-10">
      <span class="text-sm text-gray-400 dark:text-gray-600">No active connections</span>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
          <tr>
            <th class="text-left px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">Name</th>
            <th class="text-left px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">IP</th>
            <th class="text-right px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">Subs</th>
            <th class="text-right px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">Msgs In</th>
            <th class="text-right px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">Msgs Out</th>
            <th class="text-right px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">Bytes In</th>
            <th class="text-right px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">Bytes Out</th>
            <th class="text-right px-4 py-2.5 text-xs font-medium text-gray-500 dark:text-gray-400">Uptime</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="conn in connections"
            :key="conn.cid"
            class="border-b border-gray-100 dark:border-gray-800/50 hover:bg-gray-50 dark:hover:bg-gray-800/40"
          >
            <td class="px-4 py-2.5 font-mono text-xs text-gray-800 dark:text-gray-200">
              {{ conn.name || '—' }}
            </td>
            <td class="px-4 py-2.5 font-mono text-xs text-gray-600 dark:text-gray-400">
              {{ conn.ip }}:{{ conn.port }}
            </td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ conn.subs }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ conn.in_msgs.toLocaleString() }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ conn.out_msgs.toLocaleString() }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ formatBytes(conn.in_bytes) }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ formatBytes(conn.out_bytes) }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-500 dark:text-gray-500">{{ conn.uptime }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
