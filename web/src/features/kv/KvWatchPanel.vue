<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useKvWatch, type KvWatchEntry } from '../../composables/useKvWatch'

const props = defineProps<{
  bucket: string
}>()

const BUFFER_OPTIONS = [
  { label: '50', value: 50 },
  { label: '100', value: 100 },
  { label: '500', value: 500 },
  { label: '1000', value: 1000 },
]

const bufferSize = ref(100)
const keyFilter = ref('>')
const selectedEntry = ref<KvWatchEntry | null>(null)
const format = ref<'json' | 'raw'>('json')
const copyFeedback = ref(false)

const { entries, isWatching, watch: startWatch, stop, clear } = useKvWatch(500)

function startWatching() {
  startWatch(props.bucket, keyFilter.value || undefined)
}

function stopWatching() {
  stop()
  selectedEntry.value = null
}

function clearEntries() {
  clear()
  selectedEntry.value = null
}

function selectEntry(entry: KvWatchEntry) {
  if (selectedEntry.value === entry) {
    selectedEntry.value = null
  } else {
    selectedEntry.value = entry
    // default to json if parseable
    if (entry.valueText) {
      try {
        JSON.parse(entry.valueText)
        format.value = 'json'
      } catch {
        format.value = 'raw'
      }
    } else {
      format.value = 'raw'
    }
  }
}

function formatTime(iso: string): string {
  if (!iso) return '-'
  try {
    const d = new Date(iso)
    const day = d.getDate().toString().padStart(2, '0')
    const month = d.toLocaleString('en-US', { month: 'short' })
    const year = d.getFullYear()
    const hh = d.getHours().toString().padStart(2, '0')
    const mm = d.getMinutes().toString().padStart(2, '0')
    const ss = d.getSeconds().toString().padStart(2, '0')
    const ms = d.getMilliseconds().toString().padStart(3, '0')
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
    return `${day} ${month} ${year} at ${hh}:${mm}:${ss}.${ms} ${tz}`
  } catch {
    return iso
  }
}

function formatBytes(bytes: number): string {
  if (!bytes && bytes !== 0) return '-'
  if (bytes === 0) return '0 B'
  if (bytes < 1024) return `${bytes.toFixed(2)} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

function operationClass(op: string): string {
  const upper = op?.toUpperCase()
  if (upper === 'DEL' || upper === 'DELETE' || upper === 'PURGE') {
    return 'text-red-500 dark:text-red-400'
  }
  return 'text-green-600 dark:text-green-400'
}

function operationLabel(op: string): string {
  const upper = op?.toUpperCase()
  if (upper === 'DEL') return 'DELETE'
  return upper || 'PUT'
}

const formattedValue = computed(() => {
  const entry = selectedEntry.value
  if (!entry?.valueText) return entry?.valueText ?? ''
  if (format.value === 'json') {
    try {
      return JSON.stringify(JSON.parse(entry.valueText), null, 2)
    } catch {
      return entry.valueText
    }
  }
  return entry.valueText
})

const valueLines = computed(() => formattedValue.value.split('\n'))


async function copyValue() {
  try {
    await navigator.clipboard.writeText(formattedValue.value)
    copyFeedback.value = true
    setTimeout(() => { copyFeedback.value = false }, 1500)
  } catch {}
}

onUnmounted(() => {
  stop()
})
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Toolbar -->
    <div class="flex items-center gap-3 px-4 py-2.5 border-b border-gray-200 dark:border-gray-800 shrink-0 flex-wrap">
      <!-- Keys filter -->
      <div class="flex items-center gap-2">
        <label class="text-xs text-gray-500 dark:text-gray-400 shrink-0">Keys</label>
        <input
          v-model="keyFilter"
          type="text"
          placeholder="e.g. config.>"
          :disabled="isWatching"
          class="h-7 px-2 text-xs rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-gray-800 dark:text-gray-200 placeholder-gray-400 dark:placeholder-gray-600 disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-1 focus:ring-emerald-500 w-36"
        />
      </div>

      <!-- Buffer size -->
      <div class="flex items-center gap-2">
        <label class="text-xs text-gray-500 dark:text-gray-400 shrink-0">Buffer Size</label>
        <select
          v-model="bufferSize"
          :disabled="isWatching"
          class="h-7 px-2 text-xs rounded border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-gray-800 dark:text-gray-200 disabled:opacity-50 disabled:cursor-not-allowed focus:outline-none focus:ring-1 focus:ring-emerald-500"
        >
          <option v-for="opt in BUFFER_OPTIONS" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>
      </div>

      <!-- Spacer -->
      <div class="flex-1" />

      <!-- Start / Stop button -->
      <button
        v-if="!isWatching"
        class="flex items-center gap-1.5 h-7 px-3 text-xs rounded-md bg-green-600 text-white hover:bg-green-700 font-medium"
        @click="startWatching"
      >
        <!-- Play icon -->
        <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M8 5v14l11-7z"/>
        </svg>
        Start
      </button>
      <button
        v-else
        class="flex items-center gap-1.5 h-7 px-3 text-xs rounded-md bg-red-600 text-white hover:bg-red-700 font-medium"
        @click="stopWatching"
      >
        <!-- Stop icon -->
        <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M6 6h12v12H6z"/>
        </svg>
        Stop
      </button>

      <!-- Refresh (restart) -->
      <button
        class="p-1.5 rounded text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
        title="Restart watch"
        :disabled="!isWatching"
        @click="startWatching"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
        </svg>
      </button>

      <!-- Clear -->
      <button
        class="p-1.5 rounded text-gray-400 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-800"
        title="Clear changes"
        @click="clearEntries"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
        </svg>
      </button>
    </div>

    <!-- Main content: split when entry selected -->
    <div class="flex-1 overflow-hidden flex flex-col">

      <!-- Entry list -->
      <div
        class="overflow-auto"
        :class="selectedEntry ? 'flex-none max-h-64' : 'flex-1'"
      >
        <!-- Not watching, no entries -->
        <div
          v-if="!isWatching && entries.length === 0"
          class="flex items-center justify-center h-40"
        >
          <div class="text-center">
            <svg class="w-8 h-8 text-gray-300 dark:text-gray-700 mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
              <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.964-7.178Z"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/>
            </svg>
            <p class="text-sm text-gray-500 dark:text-gray-400">Click Start to watch for changes</p>
          </div>
        </div>

        <!-- Watching, no entries yet -->
        <div
          v-else-if="isWatching && entries.length === 0"
          class="flex items-center justify-center h-40"
        >
          <div class="text-center">
            <div class="w-5 h-5 rounded-full border-2 border-green-500 border-t-transparent animate-spin mx-auto mb-2"></div>
            <p class="text-sm text-gray-500 dark:text-gray-400">Watching for changes...</p>
          </div>
        </div>

        <!-- Table -->
        <table v-else class="w-full text-sm">
          <thead class="sticky top-0 bg-white dark:bg-gray-950 z-10">
            <tr class="border-b border-gray-200 dark:border-gray-800">
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 whitespace-nowrap">Time</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Operation</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Key</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Revision</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Size</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(entry, idx) in [...entries].reverse()"
              :key="`${entry.revision}-${idx}`"
              class="border-b border-gray-100 dark:border-gray-800/50 cursor-pointer"
              :class="selectedEntry === entry
                ? 'bg-emerald-50 dark:bg-emerald-950/40'
                : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'"
              @click="selectEntry(entry)"
            >
              <td class="px-4 py-2 text-xs text-gray-500 dark:text-gray-400 font-mono whitespace-nowrap">{{ formatTime(entry.receivedAt) }}</td>
              <td class="px-4 py-2 text-xs font-mono font-semibold" :class="operationClass(entry.operation)">{{ operationLabel(entry.operation) }}</td>
              <td class="px-4 py-2 text-xs font-mono text-gray-800 dark:text-gray-200">{{ entry.key }}</td>
              <td class="px-4 py-2 text-xs font-mono text-gray-600 dark:text-gray-400">{{ entry.revision }}</td>
              <td class="px-4 py-2 text-xs font-mono text-gray-600 dark:text-gray-400">{{ formatBytes(entry.size) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Detail panel (slide in from bottom when entry selected) -->
      <div
        v-if="selectedEntry"
        class="flex-1 overflow-hidden flex flex-col border-t border-gray-200 dark:border-gray-800"
      >
        <!-- Detail header / breadcrumb -->
        <div class="flex items-center gap-2 px-4 py-2.5 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <!-- Breadcrumb -->
          <div class="flex items-center gap-1.5 flex-1 min-w-0">
            <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"/>
            </svg>
            <span class="text-sm font-medium text-gray-800 dark:text-gray-200 font-mono truncate">{{ selectedEntry.key }}</span>
            <span class="text-xs text-gray-400 dark:text-gray-600 shrink-0">({{ selectedEntry.revision }})</span>
            <svg class="w-3 h-3 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
            </svg>
            <span class="text-xs text-gray-500 dark:text-gray-400 font-mono shrink-0">{{ bucket }}</span>
          </div>

          <!-- Detail actions -->
          <div class="flex items-center gap-1 shrink-0">
            <!-- Copy -->
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
              :class="copyFeedback ? 'text-green-500' : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
              title="Copy value"
              @click="copyValue"
            >
              <svg v-if="!copyFeedback" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184"/>
              </svg>
              <svg v-else class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5"/>
              </svg>
            </button>
            <!-- Close -->
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              title="Close detail"
              @click="selectedEntry = null"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Metadata row -->
        <div class="flex items-center gap-5 px-4 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0 flex-wrap">
          <div class="flex flex-col gap-0.5">
            <span class="text-xs text-gray-400 dark:text-gray-600">Key</span>
            <span class="text-xs font-mono text-gray-700 dark:text-gray-300">{{ selectedEntry.key }}</span>
          </div>
          <div class="flex flex-col gap-0.5">
            <span class="text-xs text-gray-400 dark:text-gray-600">Operation</span>
            <span class="text-xs font-mono font-semibold" :class="operationClass(selectedEntry.operation)">{{ operationLabel(selectedEntry.operation) }}</span>
          </div>
          <div class="flex flex-col gap-0.5">
            <span class="text-xs text-gray-400 dark:text-gray-600">Created</span>
            <span class="text-xs font-mono text-gray-700 dark:text-gray-300">{{ formatTime(selectedEntry.created) }}</span>
          </div>
          <div class="flex flex-col gap-0.5">
            <span class="text-xs text-gray-400 dark:text-gray-600">Value Size</span>
            <span class="text-xs font-mono text-gray-700 dark:text-gray-300">{{ formatBytes(selectedEntry.size) }}</span>
          </div>
          <div class="flex flex-col gap-0.5">
            <span class="text-xs text-gray-400 dark:text-gray-600">Revision</span>
            <span class="text-xs font-mono text-gray-700 dark:text-gray-300">{{ selectedEntry.revision }}</span>
          </div>
          <div class="flex flex-col gap-0.5">
            <span class="text-xs text-gray-400 dark:text-gray-600">KV Bucket</span>
            <span class="text-xs font-mono text-gray-700 dark:text-gray-300">{{ bucket }}</span>
          </div>
        </div>

        <!-- Format toggle + value -->
        <div class="flex items-center gap-2 px-4 py-1.5 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <div class="flex rounded-md overflow-hidden border border-gray-300 dark:border-gray-700">
            <button
              type="button"
              class="px-2.5 py-1 text-xs"
              :class="format === 'json' ? 'bg-emerald-600 text-white' : 'bg-white dark:bg-gray-900 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
              @click="format = 'json'"
            >JSON</button>
            <button
              type="button"
              class="px-2.5 py-1 text-xs border-l border-gray-300 dark:border-gray-700"
              :class="format === 'raw' ? 'bg-emerald-600 text-white' : 'bg-white dark:bg-gray-900 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
              @click="format = 'raw'"
            >Raw</button>
          </div>
          <span class="text-xs text-gray-400 dark:text-gray-600">{{ valueLines.length }} lines</span>
        </div>

        <!-- Value display -->
        <div class="flex-1 overflow-auto bg-gray-50 dark:bg-gray-900/50">
          <div v-if="!selectedEntry.valueText" class="px-4 py-3 text-xs text-gray-400 dark:text-gray-600 italic">
            No value ({{ operationLabel(selectedEntry.operation) }} operation)
          </div>
          <div v-else class="flex min-w-0 h-full">
            <!-- Line numbers -->
            <div class="select-none px-3 py-3 text-right shrink-0 border-r border-gray-200 dark:border-gray-800">
              <div
                v-for="(_, i) in valueLines"
                :key="i"
                class="text-xs font-mono leading-5 text-gray-400 dark:text-gray-600"
              >{{ i + 1 }}</div>
            </div>
            <!-- Code -->
            <pre class="flex-1 px-3 py-3 text-xs font-mono leading-5 text-gray-800 dark:text-gray-200 whitespace-pre overflow-x-auto">{{ formattedValue }}</pre>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>
