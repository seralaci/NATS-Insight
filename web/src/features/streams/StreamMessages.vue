<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useStreamsStore } from '../../stores/streams'
import { useTail } from '../../composables/useTail'
import type { StreamMessage } from '../../lib/api'
import type { TailMessage } from '../../composables/useTail'

const props = defineProps<{
  streamName: string
}>()

const emit = defineEmits<{
  selectMessage: [msg: StreamMessage]
  'open-inspector': []
}>()

const streamsStore = useStreamsStore()

const mode = ref<'realtime' | 'history'>('history')

// ── Filter state ──────────────────────────────────────────────────────────────

const filters = ref<{
  subject?: string
  startSeq?: number
  startDate?: string
}>({})

const showFilterMenu = ref(false)
const editingFilter = ref<'subject' | 'startSeq' | 'startDate' | null>(null)
const filterInput = ref('')
const filterDateInput = ref('')
const filterTimeInput = ref('')

const filterMenuRef = ref<HTMLElement | null>(null)
const filterBtnRef = ref<HTMLElement | null>(null)

const hasActiveFilters = computed(() =>
  !!filters.value.subject || !!filters.value.startSeq || !!filters.value.startDate
)

function openFilterMenu() {
  showFilterMenu.value = !showFilterMenu.value
  if (!showFilterMenu.value) editingFilter.value = null
}

function selectFilterType(type: 'subject' | 'startSeq' | 'startDate') {
  editingFilter.value = type
  if (type === 'startDate') {
    if (filters.value.startDate) {
      const d = new Date(filters.value.startDate)
      filterDateInput.value = d.toISOString().slice(0, 10)
      filterTimeInput.value = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}.${String(d.getMilliseconds()).padStart(3, '0')}`
    } else {
      const now = new Date()
      filterDateInput.value = now.toISOString().slice(0, 10)
      filterTimeInput.value = now.toTimeString().slice(0, 8) + '.' + String(now.getMilliseconds()).padStart(3, '0')
    }
  } else {
    filterInput.value =
      type === 'subject' ? (filters.value.subject ?? '') :
      (filters.value.startSeq != null ? String(filters.value.startSeq) : '')
  }
  showFilterMenu.value = false
}

function saveFilter() {
  if (!editingFilter.value) return
  const type = editingFilter.value
  const val = String(filterInput.value).trim()
  if (type === 'subject') {
    filters.value = { ...filters.value, subject: val || undefined }
  } else if (type === 'startSeq') {
    const n = parseInt(val)
    filters.value = { ...filters.value, startSeq: val && !isNaN(n) ? n : undefined }
  } else if (type === 'startDate') {
    if (filterDateInput.value && filterTimeInput.value) {
      const isoStr = new Date(`${filterDateInput.value}T${filterTimeInput.value}`).toISOString()
      filters.value = { ...filters.value, startDate: isoStr }
    } else {
      filters.value = { ...filters.value, startDate: undefined }
    }
  }
  editingFilter.value = null
  filterInput.value = ''
  filterDateInput.value = ''
  filterTimeInput.value = ''
  if (mode.value === 'realtime') {
    if (type === 'subject') {
      restartRealtime()
    } else {
      switchToHistory()
    }
  } else {
    refresh()
  }
}

function cancelFilter() {
  editingFilter.value = null
  filterInput.value = ''
  filterDateInput.value = ''
  filterTimeInput.value = ''
}

function removeFilter(key: 'subject' | 'startSeq' | 'startDate') {
  const next = { ...filters.value }
  delete next[key]
  filters.value = next
  if (mode.value === 'history') {
    refresh()
  }
}

function clearAllFilters() {
  filters.value = {}
  if (mode.value === 'history') {
    refresh()
  }
}

// Close filter menu when clicking outside
function onDocClick(e: MouseEvent) {
  if (!showFilterMenu.value) return
  const target = e.target as Node
  if (filterMenuRef.value?.contains(target) || filterBtnRef.value?.contains(target)) return
  showFilterMenu.value = false
}

onMounted(() => {
  document.addEventListener('mousedown', onDocClick)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onDocClick)
  stopRealtime()
})

const PAGE_SIZE = 50

const stream = computed(() => streamsStore.selectedStream)

// ── Realtime ──────────────────────────────────────────────────────────────────

const tail = useTail(1000)
const realtimeStarted = ref(false)
const listContainer = ref<HTMLElement | null>(null)
const autoScroll = ref(true)

const realtimeSubject = computed(() => {
  if (filters.value.subject) return filters.value.subject
  const subjects = stream.value?.config.subjects ?? []
  if (subjects.length === 0) return '>'
  if (subjects.length === 1) return subjects[0]
  return '>'
})

function startRealtime() {
  realtimeStarted.value = true
  autoScroll.value = true
  tail.start(realtimeSubject.value)
}

function stopRealtime() {
  tail.stop()
  realtimeStarted.value = false
}

function restartRealtime() {
  if (realtimeStarted.value) {
    tail.stop()
    tail.clear()
    tail.start(realtimeSubject.value)
  }
}

function switchToHistory() {
  stopRealtime()
  mode.value = 'history'
  refresh()
}

function selectRealtimeMessage(msg: TailMessage) {
  const streamMsg: StreamMessage = {
    sequence: 0,
    subject: msg.subject,
    time: msg.receivedAt,
    size: msg.size,
    data: msg.data,
    dataText: msg.dataText,
    headers: msg.headers,
  }
  emit('selectMessage', streamMsg)
}

// Auto-scroll when new messages arrive
watch(
  () => tail.messages.value,
  async () => {
    if (!autoScroll.value) return
    await nextTick()
    const el = listContainer.value
    if (el) el.scrollTop = el.scrollHeight
  },
  { flush: 'post' }
)

function onRealtimeScroll() {
  const el = listContainer.value
  if (!el) return
  autoScroll.value = el.scrollHeight - el.scrollTop - el.clientHeight < 80
}

// Stop tail when switching away from realtime mode
watch(mode, (newMode) => {
  if (newMode !== 'realtime') {
    stopRealtime()
  }
})

// ── History ───────────────────────────────────────────────────────────────────

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`
}

function formatDate(iso: string) {
  if (!iso) return '-'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function formatTime(iso: string) {
  if (!iso) return '-'
  try {
    const d = new Date(iso)
    const hh = String(d.getHours()).padStart(2, '0')
    const mm = String(d.getMinutes()).padStart(2, '0')
    const ss = String(d.getSeconds()).padStart(2, '0')
    const ms = String(d.getMilliseconds()).padStart(3, '0')
    return `${hh}:${mm}:${ss}.${ms}`
  } catch {
    return iso
  }
}

function payloadPreview(data: string | undefined, dataText: string | undefined): string {
  const text = dataText || data || ''
  if (!text) return '(empty)'
  if (text.length > 80) return text.slice(0, 80) + '...'
  return text
}

function headerCount(msg: StreamMessage): number {
  if (!msg.headers) return 0
  return Object.keys(msg.headers).length
}

function realtimeHeaderCount(msg: TailMessage): number {
  if (!msg.headers) return 0
  return Object.keys(msg.headers).length
}

async function refresh() {
  const opts: any = { limit: PAGE_SIZE }
  if (filters.value.subject) opts.subject = filters.value.subject
  if (filters.value.startSeq) opts.startSeq = filters.value.startSeq
  if (filters.value.startDate) opts.startDate = filters.value.startDate
  await streamsStore.fetchMessages(props.streamName, opts)
}

async function loadMore() {
  if (streamsStore.messages.length === 0) return
  const lastSeq = streamsStore.messages[streamsStore.messages.length - 1].sequence
  if (lastSeq <= streamsStore.messagesFirstSeq) return
  const opts: any = { limit: PAGE_SIZE, startSeq: streamsStore.messagesFirstSeq }
  if (filters.value.subject) opts.subject = filters.value.subject
  await streamsStore.fetchMoreMessages(props.streamName, opts)
}

// Auto-load messages on mount (history mode) and when stream prop changes
onMounted(() => {
  if (mode.value === 'history') {
    refresh()
  }
})

watch(
  () => props.streamName,
  (newName) => {
    if (newName && mode.value === 'history') {
      filters.value = {}
      editingFilter.value = null
      refresh()
    }
  }
)
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Stats bar -->
    <div v-if="stream" class="flex items-center gap-4 px-4 py-2 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/40 text-xs text-gray-500 dark:text-gray-400 shrink-0 flex-wrap">
      <span>Messages <strong class="text-gray-700 dark:text-gray-300 font-mono">{{ stream.state.messages }}</strong></span>
      <span>Consumers <strong class="text-gray-700 dark:text-gray-300 font-mono">{{ stream.state.consumer_count }}</strong></span>
      <span>Size <strong class="text-gray-700 dark:text-gray-300 font-mono">{{ formatBytes(stream.state.bytes) }}</strong></span>
      <span>Messages/s <strong class="text-gray-700 dark:text-gray-300 font-mono">0</strong></span>
      <span>Bytes/s <strong class="text-gray-700 dark:text-gray-300 font-mono">0.00 B</strong></span>
      <span>Cluster <strong class="text-gray-700 dark:text-gray-300 font-mono">---</strong></span>
    </div>

    <!-- Toolbar row -->
    <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0 flex-wrap">

      <!-- Eye icon — open message inspector -->
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
        title="Message Inspector"
        @click="$emit('open-inspector')"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z"/>
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/>
        </svg>
      </button>

      <!-- Filter dropdown button -->
      <div class="relative">
        <button
          ref="filterBtnRef"
          class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
          :class="hasActiveFilters || showFilterMenu || editingFilter
            ? 'text-emerald-500'
            : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
          title="Add filter"
          @click="openFilterMenu"
        >
          <!-- Funnel / filter icon -->
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 3c2.755 0 5.455.232 8.083.678.533.09.917.556.917 1.096v1.044a2.25 2.25 0 0 1-.659 1.591l-5.432 5.432a2.25 2.25 0 0 0-.659 1.591v2.927a2.25 2.25 0 0 1-1.244 2.013L9.75 21v-6.568a2.25 2.25 0 0 0-.659-1.591L3.659 7.409A2.25 2.25 0 0 1 3 5.818V4.774c0-.54.384-1.006.917-1.096A48.32 48.32 0 0 1 12 3Z"/>
          </svg>
        </button>

        <!-- Filter dropdown menu -->
        <div
          v-if="showFilterMenu"
          ref="filterMenuRef"
          class="absolute left-0 top-full mt-1 w-44 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 py-1"
        >
          <p class="px-3 pt-1 pb-1 text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider">Filters</p>
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            @click="selectFilterType('subject')"
          >
            <!-- Tag icon -->
            <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 0 0 5.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 0 0 9.568 3Z"/>
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 6h.008v.008H6V6Z"/>
            </svg>
            Subject
          </button>
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            @click="selectFilterType('startSeq')"
          >
            <!-- Hash icon -->
            <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5-3.9 19.5m-2.1-19.5-3.9 19.5"/>
            </svg>
            Start Sequence
          </button>
          <button
            class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            @click="selectFilterType('startDate')"
          >
            <!-- Calendar icon -->
            <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5"/>
            </svg>
            Start Date
          </button>
        </div>
      </div>

      <!-- Divider -->
      <span class="text-gray-300 dark:text-gray-700 select-none">|</span>

      <!-- Count indicator -->
      <span class="text-xs text-gray-500 dark:text-gray-400">
        <template v-if="mode === 'realtime'">
          <span
            class="inline-block w-1.5 h-1.5 rounded-full mr-1 align-middle"
            :class="tail.isRunning.value ? 'bg-emerald-500' : 'bg-gray-400 dark:bg-gray-600'"
          ></span>
          {{ tail.messageCount.value }} live messages
        </template>
        <template v-else>
          {{ streamsStore.messages.length }} fetched messages
        </template>
      </span>

      <!-- Active filter chips -->
      <template v-if="hasActiveFilters">
        <!-- Clear all chip -->
        <button
          class="flex items-center gap-1 px-2 py-0.5 text-xs bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700"
          title="Clear all filters"
          @click="clearAllFilters"
        >
          <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
          </svg>
        </button>

        <!-- Subject chip -->
        <span
          v-if="filters.subject"
          class="flex items-center gap-1 px-2 py-0.5 text-xs bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300 rounded-full"
        >
          <svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 0 0 5.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 0 0 9.568 3Z"/>
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 6h.008v.008H6V6Z"/>
          </svg>
          {{ filters.subject }}
          <button class="ml-0.5 hover:text-emerald-900 dark:hover:text-emerald-100" @click="removeFilter('subject')">
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </span>

        <!-- Start Sequence chip -->
        <span
          v-if="filters.startSeq != null"
          class="flex items-center gap-1 px-2 py-0.5 text-xs bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300 rounded-full"
        >
          <svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5-3.9 19.5m-2.1-19.5-3.9 19.5"/>
          </svg>
          {{ filters.startSeq }}
          <button class="ml-0.5 hover:text-emerald-900 dark:hover:text-emerald-100" @click="removeFilter('startSeq')">
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </span>

        <!-- Start Date chip -->
        <span
          v-if="filters.startDate"
          class="flex items-center gap-1 px-2 py-0.5 text-xs bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300 rounded-full"
        >
          <svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5"/>
          </svg>
          {{ formatDate(filters.startDate!) }}
          <button class="ml-0.5 hover:text-emerald-900 dark:hover:text-emerald-100" @click="removeFilter('startDate')">
            <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </span>
      </template>

      <!-- Spacer -->
      <div class="flex-1"></div>

      <!-- Refresh (history only) -->
      <button
        v-if="mode === 'history'"
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-40"
        :disabled="streamsStore.messagesLoading"
        title="Refresh"
        @click="refresh"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
        </svg>
      </button>

      <!-- Realtime controls: Pause/Resume + Clear -->
      <template v-if="mode === 'realtime' && realtimeStarted">
        <button
          class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
          :class="tail.isPaused.value ? 'text-amber-500' : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'"
          :title="tail.isPaused.value ? 'Resume' : 'Pause'"
          @click="tail.isPaused.value ? tail.resume() : tail.pause()"
        >
          <svg v-if="!tail.isPaused.value" class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
          </svg>
          <svg v-else class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M8 5v14l11-7z"/>
          </svg>
        </button>

        <button
          class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          title="Clear messages"
          @click="tail.clear()"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
          </svg>
        </button>
      </template>

      <!-- Realtime / History toggle -->
      <div class="flex rounded-md overflow-hidden border border-gray-300 dark:border-gray-700">
        <button
          type="button"
          class="px-3 py-1 text-xs font-medium"
          :class="mode === 'realtime'
            ? 'bg-emerald-600 text-white'
            : 'bg-white dark:bg-gray-900 text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
          @click="mode = 'realtime'"
        >Realtime</button>
        <button
          type="button"
          class="px-3 py-1 text-xs font-medium border-l border-gray-300 dark:border-gray-700"
          :class="mode === 'history'
            ? 'bg-emerald-600 text-white'
            : 'bg-white dark:bg-gray-900 text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
          @click="switchToHistory"
        >History</button>
      </div>
    </div>

    <!-- Filter inline editor (shown below toolbar when editing a filter) -->
    <div
      v-if="editingFilter"
      class="flex items-end gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/40 shrink-0"
    >
      <template v-if="editingFilter === 'startDate'">
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Date</label>
          <input
            v-model="filterDateInput"
            type="date"
            class="px-3 py-1.5 text-xs border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-emerald-500 w-36"
            @keydown.escape="cancelFilter"
          />
        </div>
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">Time</label>
          <input
            v-model="filterTimeInput"
            type="text"
            placeholder="HH:MM:SS.mmm"
            class="px-3 py-1.5 text-xs border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-emerald-500 w-32"
            @keydown.enter="saveFilter"
            @keydown.escape="cancelFilter"
          />
        </div>
        <p class="text-xs text-gray-400 dark:text-gray-600 self-end mb-1">Your local timezone gets used</p>
      </template>
      <template v-else>
        <div class="flex flex-col gap-1">
          <label class="text-xs font-medium text-gray-600 dark:text-gray-400">
            <template v-if="editingFilter === 'subject'">Subject</template>
            <template v-else>Start Sequence</template>
          </label>
          <input
            v-if="editingFilter === 'startSeq'"
            v-model="filterInput"
            type="number"
            placeholder="Start Sequence"
            class="px-3 py-1.5 text-xs border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-emerald-500 w-44"
            @keydown.enter="saveFilter"
            @keydown.escape="cancelFilter"
          />
          <input
            v-else
            v-model="filterInput"
            type="text"
            placeholder="Subject"
            class="px-3 py-1.5 text-xs border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-emerald-500 w-44"
            @keydown.enter="saveFilter"
            @keydown.escape="cancelFilter"
          />
        </div>
      </template>
      <button
        class="px-4 py-1.5 text-xs bg-emerald-600 text-white rounded-md hover:bg-emerald-700 mb-0.5"
        @click="saveFilter"
      >Save</button>
      <button
        class="px-3 py-1.5 text-xs border border-gray-300 dark:border-gray-700 rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 mb-0.5"
        @click="cancelFilter"
      >Cancel</button>
    </div>

    <!-- ── REALTIME MODE ──────────────────────────────────────────────────── -->
    <template v-if="mode === 'realtime'">
      <!-- Not started yet: prompt to connect -->
      <div v-if="!realtimeStarted" class="flex-1 flex flex-col items-center justify-center gap-3 text-center px-6">
        <svg class="w-10 h-10 text-emerald-400 dark:text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z"/>
        </svg>
        <p class="text-sm font-medium text-gray-700 dark:text-gray-300">Live message stream</p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          Subscribe to
          <span class="font-mono text-emerald-600 dark:text-emerald-400">{{ realtimeSubject }}</span>
          and watch messages arrive in real time.
        </p>
        <button
          class="mt-2 px-4 py-1.5 text-xs bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
          @click="startRealtime"
        >Start Listening</button>
      </div>

      <!-- Realtime table -->
      <div
        v-else
        ref="listContainer"
        class="flex-1 overflow-auto"
        @scroll="onRealtimeScroll"
      >
        <!-- Connecting / empty state -->
        <div v-if="tail.messages.value.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
          <template v-if="!tail.isRunning.value">
            <svg class="w-5 h-5 text-gray-400 dark:text-gray-600 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
            </svg>
            <span class="text-xs text-gray-400 dark:text-gray-500">Connecting...</span>
          </template>
          <template v-else>
            <svg class="w-10 h-10 text-gray-300 dark:text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 13.5l10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75z"/>
            </svg>
            <span class="text-xs text-gray-500 dark:text-gray-400">
              Listening on <span class="font-mono text-emerald-600 dark:text-emerald-400">{{ realtimeSubject }}</span> — waiting for messages
            </span>
          </template>
        </div>

        <!-- Message table -->
        <table v-else class="w-full text-sm">
          <thead class="sticky top-0 bg-white dark:bg-gray-950 border-b border-gray-200 dark:border-gray-800">
            <tr>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-44">Subject</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-36">Received</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-20">Size</th>
              <th class="text-right px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-20">Headers</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Payload</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="msg in tail.messages.value"
              :key="msg.id"
              class="border-b border-gray-100 dark:border-gray-800/50 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/40"
              @click="selectRealtimeMessage(msg)"
            >
              <td class="px-4 py-2 font-mono text-xs text-emerald-600 dark:text-emerald-400 truncate max-w-0 w-44">
                <span class="block truncate">{{ msg.subject }}</span>
              </td>
              <td class="px-4 py-2 text-xs text-gray-500 dark:text-gray-400 font-mono">{{ formatTime(msg.receivedAt) }}</td>
              <td class="px-4 py-2 text-xs font-mono text-gray-500 dark:text-gray-400">{{ formatBytes(msg.size) }}</td>
              <td class="px-4 py-2 text-xs font-mono text-right text-gray-500 dark:text-gray-400">{{ realtimeHeaderCount(msg) }}</td>
              <td class="px-4 py-2 text-xs font-mono text-gray-600 dark:text-gray-400 truncate max-w-0">
                <span class="block truncate">{{ payloadPreview(msg.data, msg.dataText) }}</span>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Paused banner -->
        <div v-if="tail.isPaused.value" class="sticky bottom-0 flex items-center justify-center gap-2 py-1.5 bg-amber-50 dark:bg-amber-900/20 border-t border-amber-200 dark:border-amber-800 text-xs text-amber-700 dark:text-amber-400">
          <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
          </svg>
          Paused — {{ tail.messageCount.value - tail.messages.value.length > 0 ? `${tail.messageCount.value - tail.messages.value.length} messages dropped` : 'no new messages dropped' }}
          <button class="font-medium underline ml-1" @click="tail.resume()">Resume</button>
        </div>
      </div>
    </template>

    <!-- ── HISTORY MODE ───────────────────────────────────────────────────── -->
    <template v-else>
      <div class="flex-1 overflow-auto">
        <div v-if="streamsStore.messagesLoading && streamsStore.messages.length === 0" class="flex items-center justify-center py-20">
          <span class="text-sm text-gray-400 dark:text-gray-600">Loading messages...</span>
        </div>

        <div v-else-if="streamsStore.messages.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
          <svg class="w-10 h-10 text-gray-300 dark:text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"/>
          </svg>
          <span class="text-sm text-gray-500 dark:text-gray-400">No messages found</span>
        </div>

        <table v-else class="w-full text-sm">
          <thead class="sticky top-0 bg-white dark:bg-gray-950 border-b border-gray-200 dark:border-gray-800">
            <tr>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-24">Sequence</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-44">Subject</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-44">Received</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-20">Size</th>
              <th class="text-right px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 w-20">Headers</th>
              <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Payload</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="msg in streamsStore.messages"
              :key="msg.sequence"
              class="border-b border-gray-100 dark:border-gray-800/50 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/40"
              @click="emit('selectMessage', msg)"
            >
              <td class="px-4 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">{{ msg.sequence }}</td>
              <td class="px-4 py-2 font-mono text-xs text-emerald-600 dark:text-emerald-400 truncate max-w-0 w-44">
                <span class="block truncate">{{ msg.subject }}</span>
              </td>
              <td class="px-4 py-2 text-xs text-gray-500 dark:text-gray-400">{{ formatDate(msg.time) }}</td>
              <td class="px-4 py-2 text-xs font-mono text-gray-500 dark:text-gray-400">{{ formatBytes(msg.size) }}</td>
              <td class="px-4 py-2 text-xs font-mono text-right text-gray-500 dark:text-gray-400">{{ headerCount(msg) }}</td>
              <td class="px-4 py-2 text-xs font-mono text-gray-600 dark:text-gray-400 truncate max-w-0">
                <span class="block truncate">{{ payloadPreview(msg.data, msg.dataText) }}</span>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Load More / end of messages -->
        <div v-if="streamsStore.messages.length > 0" class="flex flex-col items-center py-6 gap-2">
          <template v-if="streamsStore.messages.length < streamsStore.messagesTotal">
            <button
              class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-50"
              :disabled="streamsStore.messagesLoading"
              @click="loadMore"
            >
              {{ streamsStore.messagesLoading ? 'Loading...' : 'Try to load more messages' }}
            </button>
          </template>
          <template v-else>
            <p class="text-xs text-gray-400 dark:text-gray-600">You've reached the end of the messages.</p>
            <button
              class="text-xs text-emerald-500 hover:underline"
              :disabled="streamsStore.messagesLoading"
              @click="loadMore"
            >Try to load more messages</button>
          </template>
        </div>
      </div>
    </template>
  </div>
</template>
