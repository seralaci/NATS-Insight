<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { streamsApi, type StreamMessage } from '../../lib/api'
import { useStreamsStore } from '../../stores/streams'

const props = defineProps<{
  streamName: string
}>()

const emit = defineEmits<{
  back: []
}>()

const streamsStore = useStreamsStore()

const mode = ref<'sequence' | 'subject'>('sequence')
const searchInput = ref('')
const currentMessage = ref<StreamMessage | null>(null)
const error = ref('')
const loading = ref(false)
const currentSeq = ref<number | null>(null)

const showDotsMenu = ref(false)
const dotsMenuRef = ref<HTMLElement | null>(null)
const dotsBtnRef = ref<HTMLElement | null>(null)

const activeTab = ref<'payload' | 'headers'>('payload')
const format = ref<'json' | 'raw'>('json')
const copyFeedback = ref(false)

const streamSubjects = computed(() => {
  const subjects = streamsStore.selectedStream?.config.subjects
  if (!subjects || subjects.length === 0) return ''
  return subjects.join(', ')
})

const headerCount = computed(() => {
  if (!currentMessage.value?.headers) return 0
  return Object.keys(currentMessage.value.headers).length
})

const payloadText = computed(() => {
  return currentMessage.value?.dataText || currentMessage.value?.data || ''
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

const canNavigatePrev = computed(() => currentSeq.value !== null && currentSeq.value > 1)
const canNavigateNext = computed(() => currentSeq.value !== null)

function toggleMode() {
  mode.value = mode.value === 'sequence' ? 'subject' : 'sequence'
  searchInput.value = ''
  currentMessage.value = null
  error.value = ''
  currentSeq.value = null
}

function onDotsDocClick(e: MouseEvent) {
  if (!showDotsMenu.value) return
  const target = e.target as Node
  if (dotsMenuRef.value?.contains(target) || dotsBtnRef.value?.contains(target)) return
  showDotsMenu.value = false
}

async function fetchBySeq(seq: number) {
  loading.value = true
  error.value = ''
  currentMessage.value = null
  try {
    const msg = await streamsApi.getMessage(props.streamName, seq)
    currentMessage.value = msg
    currentSeq.value = msg.sequence
    searchInput.value = String(msg.sequence)
    activeTab.value = 'payload'
    format.value = isJson.value ? 'json' : 'raw'
  } catch (e: any) {
    error.value = e.message || 'message not found'
    currentSeq.value = seq
  } finally {
    loading.value = false
  }
}

async function fetchBySubject(subject: string) {
  loading.value = true
  error.value = ''
  currentMessage.value = null
  try {
    const msg = await streamsApi.getLastBySubject(props.streamName, subject)
    currentMessage.value = msg
    currentSeq.value = msg.sequence
    activeTab.value = 'payload'
    format.value = isJson.value ? 'json' : 'raw'
  } catch (e: any) {
    error.value = e.message || 'message not found'
  } finally {
    loading.value = false
  }
}

async function handleSearch() {
  const val = searchInput.value.trim()
  if (!val) return
  if (mode.value === 'sequence') {
    const match = val.match(/\d+/)
    if (!match) return
    await fetchBySeq(parseInt(match[0]))
  } else {
    await fetchBySubject(val.split(/\s+/)[0])
  }
}

async function navigatePrev() {
  if (currentSeq.value === null || currentSeq.value <= 1) return
  await fetchBySeq(currentSeq.value - 1)
}

async function navigateNext() {
  if (currentSeq.value === null) return
  await fetchBySeq(currentSeq.value + 1)
}

async function refresh() {
  if (mode.value === 'sequence' && currentSeq.value !== null) {
    await fetchBySeq(currentSeq.value)
  } else if (mode.value === 'subject') {
    const val = searchInput.value.trim()
    if (val) await fetchBySubject(val.split(/\s+/)[0])
  }
}

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

onMounted(() => {
  document.addEventListener('mousedown', onDotsDocClick)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onDotsDocClick)
})

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
    <!-- Toolbar -->
    <div class="flex items-center gap-2 px-3 py-2 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-950 shrink-0">
      <!-- Mode toggle -->
      <button
        class="shrink-0 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-md px-3 py-1.5 text-xs font-medium whitespace-nowrap"
        @click="toggleMode"
      >
        <span v-if="mode === 'sequence'">&#8597; Sequence</span>
        <span v-else>&#8597; Last by Subject</span>
      </button>

      <!-- Search input -->
      <input
        v-model="searchInput"
        type="text"
        class="flex-1 min-w-0 px-3 py-1.5 text-xs border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
        :placeholder="mode === 'sequence'
          ? 'Enter sequence(s) separated by spaces'
          : `Enter subject(s) separated by spaces${streamSubjects ? '. Subjects: ' + streamSubjects : ''}`"
        @keydown.enter="handleSearch"
      />

      <!-- Nav: next (↓) -->
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-30"
        title="Next message"
        :disabled="!canNavigateNext || loading"
        @click="navigateNext"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 13.5 12 21m0 0-7.5-7.5M12 21V3"/>
        </svg>
      </button>

      <!-- Nav: prev (↑) -->
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-30"
        title="Previous message"
        :disabled="!canNavigatePrev || loading"
        @click="navigatePrev"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 10.5 12 3m0 0 7.5 7.5M12 3v18"/>
        </svg>
      </button>

      <!-- "..." menu -->
      <div class="relative">
        <button
          ref="dotsBtnRef"
          class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          title="More options"
          @click="showDotsMenu = !showDotsMenu"
        >
          <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
            <circle cx="5" cy="12" r="1.5"/>
            <circle cx="12" cy="12" r="1.5"/>
            <circle cx="19" cy="12" r="1.5"/>
          </svg>
        </button>
        <div
          v-if="showDotsMenu"
          ref="dotsMenuRef"
          class="absolute right-0 top-full mt-1 w-40 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 py-1"
        >
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            @click="showDotsMenu = false; emit('back')"
          >
            <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"/>
            </svg>
            Open Stream
          </button>
        </div>
      </div>

      <!-- Refresh -->
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-30"
        title="Refresh"
        :disabled="loading"
        @click="refresh"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
        </svg>
      </button>
    </div>

    <!-- Content area -->
    <div class="flex-1 overflow-hidden flex flex-col min-h-0">
      <!-- Empty state -->
      <div v-if="!loading && !error && !currentMessage" class="flex-1 flex items-center justify-center">
        <span class="text-sm text-gray-400 dark:text-gray-600">Enter a sequence or subject and press Enter to inspect a message</span>
      </div>

      <!-- Loading -->
      <div v-else-if="loading" class="flex-1 flex items-center justify-center">
        <span class="text-sm text-gray-400 dark:text-gray-600">Loading...</span>
      </div>

      <!-- Error -->
      <div v-else-if="error && !currentMessage" class="px-4 py-3 shrink-0">
        <div class="flex items-center gap-2 px-3 py-2 rounded-md bg-red-50 dark:bg-red-950/50 text-red-600 dark:text-red-400 border border-red-200 dark:border-red-800 text-sm">
          <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"/>
          </svg>
          {{ error }}
        </div>
      </div>

      <!-- Message detail -->
      <template v-else-if="currentMessage">
        <!-- Tabs + metadata bar -->
        <div class="border-b border-gray-200 dark:border-gray-800 shrink-0">
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
            <div class="ml-auto flex items-center gap-1 py-1">
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

          <!-- Metadata bar -->
          <div class="flex items-center gap-6 px-4 py-2 bg-gray-50 dark:bg-gray-900/40 text-xs text-gray-500 dark:text-gray-400 flex-wrap">
            <span>Sequence <strong class="text-gray-700 dark:text-gray-300 font-mono">{{ currentMessage.sequence }}</strong></span>
            <span>Subject <strong class="text-emerald-600 dark:text-emerald-400 font-mono">{{ currentMessage.subject }}</strong></span>
            <span>Received <strong class="text-gray-700 dark:text-gray-300">{{ formatDate(currentMessage.time) }}</strong></span>
            <span>Size <strong class="text-gray-700 dark:text-gray-300 font-mono">{{ formatBytes(currentMessage.size) }}</strong></span>
          </div>
        </div>

        <!-- Payload tab -->
        <div v-if="activeTab === 'payload'" class="flex flex-col flex-1 min-h-0">
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

          <div class="flex-1 overflow-auto bg-gray-900">
            <div v-if="!payloadText" class="flex items-center justify-center py-12">
              <span class="text-xs text-gray-500">(empty payload)</span>
            </div>
            <div v-else class="flex min-w-0 min-h-full">
              <div class="select-none px-3 py-3 text-right shrink-0 border-r border-gray-700/50">
                <div
                  v-for="(_, i) in payloadLines"
                  :key="i"
                  class="text-xs font-mono leading-5 text-gray-600"
                >{{ i + 1 }}</div>
              </div>
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
                v-for="(vals, hkey) in currentMessage.headers"
                :key="hkey"
                class="border-b border-gray-100 dark:border-gray-800/50"
              >
                <td class="px-4 py-2 font-mono text-xs text-gray-500 dark:text-gray-400">{{ hkey }}</td>
                <td class="px-4 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">{{ vals.join(', ') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </div>
</template>
