<script setup lang="ts">
import { ref, computed } from 'vue'
import type { StreamMessage } from '../../lib/api'

const props = defineProps<{
  message: StreamMessage
  streamName: string
}>()

const emit = defineEmits<{
  back: []
}>()

const activeTab = ref<'payload' | 'headers'>('payload')
const format = ref<'json' | 'raw'>('json')
const copyFeedback = ref(false)

const headerCount = computed(() => {
  if (!props.message.headers) return 0
  return Object.keys(props.message.headers).length
})

const payloadText = computed(() => {
  return props.message.dataText || props.message.data || ''
})

const formattedPayload = computed(() => {
  const text = payloadText.value
  if (!text) return ''
  if (format.value === 'json') {
    try {
      return JSON.stringify(JSON.parse(text), null, 2)
    } catch {
      return text
    }
  }
  return text
})

const payloadLines = computed(() => {
  const v = formattedPayload.value
  if (!v) return []
  return v.split('\n')
})

const isJson = computed(() => {
  const text = payloadText.value
  if (!text) return false
  try {
    JSON.parse(text)
    return true
  } catch {
    return false
  }
})

function formatDate(iso: string) {
  if (!iso) return '-'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

async function copyPayload() {
  const text = formattedPayload.value || payloadText.value
  try {
    await navigator.clipboard.writeText(text)
    copyFeedback.value = true
    setTimeout(() => { copyFeedback.value = false }, 1500)
  } catch {}
}
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header bar -->
    <div class="flex items-center gap-2 px-4 py-2.5 border-b border-gray-200 dark:border-gray-800 shrink-0">
      <button
        class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
        @click="emit('back')"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5"/>
        </svg>
        Sequence {{ message.sequence }}
      </button>
      <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
      </svg>
      <span class="text-xs font-mono font-semibold text-gray-700 dark:text-gray-300">{{ streamName }}</span>
      <div class="ml-auto flex items-center gap-1">
        <!-- Copy -->
        <button
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
          :class="copyFeedback ? 'text-green-500' : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
          title="Copy payload"
          @click="copyPayload"
        >
          <svg v-if="!copyFeedback" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184"/>
          </svg>
          <svg v-else class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Tabs + metadata bar -->
    <div class="border-b border-gray-200 dark:border-gray-800 shrink-0">
      <!-- Tabs -->
      <div class="flex items-center gap-0 px-4">
        <button
          class="px-3 py-2 text-sm border-b-2"
          :class="activeTab === 'payload'
            ? 'border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 font-medium'
            : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
          @click="activeTab = 'payload'"
        >Payload</button>
        <button
          class="px-3 py-2 text-sm border-b-2"
          :class="activeTab === 'headers'
            ? 'border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 font-medium'
            : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
          @click="activeTab = 'headers'"
        >
          Headers
          <span v-if="headerCount > 0" class="ml-1 px-1.5 py-0.5 text-[10px] rounded-full bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400">{{ headerCount }}</span>
        </button>
      </div>

      <!-- Metadata bar -->
      <div class="flex items-center gap-6 px-4 py-2 bg-gray-50 dark:bg-gray-900/40 text-xs text-gray-500 dark:text-gray-400 flex-wrap">
        <span>Sequence <strong class="text-gray-700 dark:text-gray-300 font-mono">{{ message.sequence }}</strong></span>
        <span>Subject <strong class="text-emerald-600 dark:text-emerald-400 font-mono">{{ message.subject }}</strong></span>
        <span>Received <strong class="text-gray-700 dark:text-gray-300">{{ formatDate(message.time) }}</strong></span>
        <span>Size <strong class="text-gray-700 dark:text-gray-300 font-mono">{{ formatBytes(message.size) }}</strong></span>
      </div>
    </div>

    <!-- Payload tab -->
    <div v-if="activeTab === 'payload'" class="flex flex-col flex-1 min-h-0">
      <!-- Format toolbar -->
      <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0">
        <div class="flex rounded-md overflow-hidden border border-gray-300 dark:border-gray-700">
          <button
            type="button"
            class="px-2.5 py-1 text-xs"
            :class="format === 'json' ? 'bg-emerald-600 text-white' : 'bg-white dark:bg-gray-900 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
            :disabled="!isJson"
            @click="format = 'json'"
          >JSON</button>
          <button
            type="button"
            class="px-2.5 py-1 text-xs border-l border-gray-300 dark:border-gray-700"
            :class="format === 'raw' ? 'bg-emerald-600 text-white' : 'bg-white dark:bg-gray-900 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
            @click="format = 'raw'"
          >Raw</button>
        </div>
        <span class="text-xs text-gray-400 dark:text-gray-600">{{ payloadLines.length }} lines</span>
      </div>

      <!-- Code viewer with line numbers -->
      <div class="flex-1 overflow-auto bg-gray-900">
        <div v-if="!payloadText" class="flex items-center justify-center py-12">
          <span class="text-xs text-gray-500">(empty payload)</span>
        </div>
        <div v-else class="flex min-w-0 min-h-full">
          <!-- Line numbers -->
          <div class="select-none px-3 py-3 text-right shrink-0 border-r border-gray-700/50">
            <div
              v-for="(_, i) in payloadLines"
              :key="i"
              class="text-xs font-mono leading-5 text-gray-600"
            >{{ i + 1 }}</div>
          </div>
          <!-- Code -->
          <pre class="flex-1 px-3 py-3 text-xs font-mono leading-5 text-green-400 whitespace-pre overflow-x-auto">{{ formattedPayload }}</pre>
        </div>
      </div>
    </div>

    <!-- Headers tab -->
    <div v-else-if="activeTab === 'headers'" class="flex-1 overflow-auto">
      <div v-if="headerCount === 0" class="flex items-center justify-center py-20">
        <span class="text-sm text-gray-400 dark:text-gray-600">No headers</span>
      </div>
      <table v-else class="w-full text-sm">
        <thead class="sticky top-0 bg-white dark:bg-gray-950 border-b border-gray-200 dark:border-gray-800">
          <tr>
            <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-64">Header</th>
            <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Value</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(vals, hkey) in message.headers"
            :key="hkey"
            class="border-b border-gray-100 dark:border-gray-800/50"
          >
            <td class="px-4 py-2 font-mono text-xs text-gray-500 dark:text-gray-400">{{ hkey }}</td>
            <td class="px-4 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">{{ vals.join(', ') }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
