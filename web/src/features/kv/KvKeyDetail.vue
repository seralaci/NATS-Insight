<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useKvStore } from '../../stores/kv'
import type { KvEntry } from '../../lib/api'
import KvKeyFormModal from './KvKeyFormModal.vue'

const props = defineProps<{
  bucketName: string
}>()

const kvStore = useKvStore()
const format = ref<'json' | 'raw'>('json')
const showEditModal = ref(false)

defineExpose({ openEdit: () => { showEditModal.value = true } })
const previewEntry = ref<KvEntry | null>(null)
const copyFeedback = ref(false)

const currentEntry = computed(() => previewEntry.value ?? kvStore.keyDetail)

const formattedValue = computed(() => {
  const entry = currentEntry.value
  if (!entry) return ''
  const text = entry.valueText || ''
  if (format.value === 'json') {
    try {
      return JSON.stringify(JSON.parse(text), null, 2)
    } catch {
      return text
    }
  }
  return text
})

const isJson = computed(() => {
  const text = kvStore.keyDetail?.valueText || ''
  try {
    JSON.parse(text)
    return true
  } catch {
    return false
  }
})

const valueLines = computed(() => {
  return formattedValue.value.split('\n')
})

const fileSizeDisplay = computed(() => {
  const entry = currentEntry.value
  if (!entry) return '-'
  const bytes = entry.value.length
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
})

function formatDate(iso: string) {
  if (!iso) return '-'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function historyOperationClass(op: string) {
  if (op === 'DEL' || op === 'PURGE') return 'text-red-500 dark:text-red-400'
  return 'text-green-600 dark:text-green-400'
}

function selectHistoryRevision(entry: KvEntry) {
  previewEntry.value = previewEntry.value?.revision === entry.revision ? null : entry
}

async function refresh() {
  if (kvStore.selectedKey) {
    previewEntry.value = null
    await kvStore.fetchKeyDetail(props.bucketName, kvStore.selectedKey)
  }
}

async function copyValue() {
  const text = formattedValue.value
  try {
    await navigator.clipboard.writeText(text)
    copyFeedback.value = true
    setTimeout(() => { copyFeedback.value = false }, 1500)
  } catch {}
}

async function deleteCurrentKey() {
  if (!kvStore.selectedKey) return
  if (confirm(`Delete key "${kvStore.selectedKey}"?`)) {
    await kvStore.deleteKey(props.bucketName, kvStore.selectedKey)
  }
}

watch(() => kvStore.keyDetail, () => {
  previewEntry.value = null
  if (isJson.value) format.value = 'json'
  else format.value = 'raw'
})
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Empty state -->
    <div v-if="!kvStore.selectedKey" class="flex-1 flex items-center justify-center">
      <span class="text-sm text-gray-400 dark:text-gray-600">Select a key</span>
    </div>

    <template v-else>
      <!-- Key header -->
      <div class="flex items-center gap-2 px-4 py-2.5 border-b border-gray-200 dark:border-gray-800 shrink-0">
        <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"/>
        </svg>
        <span class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate flex-1">{{ kvStore.selectedKey }}</span>

        <template v-if="kvStore.keyDetail">
          <span class="text-xs text-gray-400 dark:text-gray-600 shrink-0">
            Revision: {{ kvStore.keyDetail.revision }} &nbsp;|&nbsp;
            Size: {{ fileSizeDisplay }} &nbsp;|&nbsp;
            Created: {{ formatDate(kvStore.keyDetail.created) }}
          </span>
        </template>

        <!-- Actions -->
        <div class="flex items-center gap-1 shrink-0">
          <button
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            title="Refresh"
            @click="refresh"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
            </svg>
          </button>
          <button
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-emerald-500"
            title="Edit / New Revision"
            @click="showEditModal = true"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125"/>
            </svg>
          </button>
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
          <button
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-red-500"
            title="Delete key"
            @click="deleteCurrentKey"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="kvStore.keyDetailLoading" class="flex-1 flex items-center justify-center">
        <span class="text-sm text-gray-400 dark:text-gray-600">Loading...</span>
      </div>

      <template v-else-if="kvStore.keyDetail">
        <!-- Preview banner -->
        <div v-if="previewEntry" class="px-4 py-2 bg-amber-50 dark:bg-amber-950/30 border-b border-amber-200 dark:border-amber-800 text-xs text-amber-700 dark:text-amber-400 flex items-center gap-2 shrink-0">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"/>
          </svg>
          Viewing revision {{ previewEntry.revision }}
          <button class="ml-auto underline hover:no-underline" @click="previewEntry = null">Back to latest</button>
        </div>

        <!-- Format toggle -->
        <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0">
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
          <span class="text-xs text-gray-400 dark:text-gray-600">
            {{ valueLines.length }} lines
          </span>
        </div>

        <!-- Value display -->
        <div class="flex-none overflow-auto max-h-72 bg-gray-50 dark:bg-gray-900/50 border-b border-gray-200 dark:border-gray-800">
          <div class="flex min-w-0">
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

        <!-- History section -->
        <div class="flex-1 overflow-auto">
          <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-800">
            <span class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              History ({{ kvStore.keyHistory.length }})
            </span>
          </div>
          <div v-if="kvStore.keyHistory.length === 0" class="px-4 py-3 text-xs text-gray-400 dark:text-gray-600 italic">
            No history available
          </div>
          <table v-else class="w-full text-xs">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-800">
                <th class="text-left px-4 py-1.5 font-medium text-gray-500 dark:text-gray-400">Revision</th>
                <th class="text-left px-4 py-1.5 font-medium text-gray-500 dark:text-gray-400">Created</th>
                <th class="text-left px-4 py-1.5 font-medium text-gray-500 dark:text-gray-400">Size</th>
                <th class="text-left px-4 py-1.5 font-medium text-gray-500 dark:text-gray-400">Operation</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="entry in kvStore.keyHistory"
                :key="entry.revision"
                class="border-b border-gray-100 dark:border-gray-800/50 cursor-pointer"
                :class="[
                  previewEntry?.revision === entry.revision ? 'bg-amber-50 dark:bg-amber-950/30' : 'hover:bg-gray-50 dark:hover:bg-gray-800/50',
                ]"
                @click="selectHistoryRevision(entry)"
              >
                <td class="px-4 py-1.5 font-mono text-gray-700 dark:text-gray-300">{{ entry.revision }}</td>
                <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400">{{ formatDate(entry.created) }}</td>
                <td class="px-4 py-1.5 font-mono text-gray-500 dark:text-gray-400">{{ entry.value.length }} B</td>
                <td class="px-4 py-1.5 font-mono font-medium" :class="historyOperationClass(entry.operation)">{{ entry.operation || 'PUT' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </template>

    <!-- Edit modal -->
    <KvKeyFormModal
      v-if="showEditModal && kvStore.keyDetail"
      :bucket-name="bucketName"
      :edit-entry="kvStore.keyDetail"
      @close="showEditModal = false"
      @saved="showEditModal = false"
    />
  </div>
</template>
