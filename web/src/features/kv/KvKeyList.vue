<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useKvStore } from '../../stores/kv'
import KvKeyFormModal from './KvKeyFormModal.vue'

const props = defineProps<{
  bucketName: string
}>()

const emit = defineEmits<{
  'edit-key': [key: string]
}>()

const kvStore = useKvStore()

const searchQuery = ref('')
const showCreateModal = ref(false)
const contextMenu = ref<{ key: string; x: number; y: number } | null>(null)
const copyFeedback = ref<string | null>(null)

const filteredKeys = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return kvStore.keys
  return kvStore.keys.filter(k => k.toLowerCase().includes(q))
})

watch(searchQuery, (q) => {
  kvStore.fetchKeys(props.bucketName, q || undefined)
})

function selectKey(key: string) {
  contextMenu.value = null
  kvStore.fetchKeyDetail(props.bucketName, key)
}

function openContextMenu(event: MouseEvent, key: string) {
  event.preventDefault()
  contextMenu.value = { key, x: event.clientX, y: event.clientY }
}

function closeContextMenu() {
  contextMenu.value = null
}

async function copyText(text: string, label: string) {
  try {
    await navigator.clipboard.writeText(text)
    copyFeedback.value = label
    setTimeout(() => { copyFeedback.value = null }, 1500)
  } catch {}
}

async function copyKey(key: string) {
  closeContextMenu()
  await copyText(key, 'Key copied')
}

async function copyValue(key: string) {
  closeContextMenu()
  try {
    const entry = await import('../../lib/api').then(m => m.kvApi.getKey(props.bucketName, key))
    await copyText(entry.valueText || entry.value, 'Value copied')
  } catch {}
}

function editKey(key: string) {
  closeContextMenu()
  kvStore.fetchKeyDetail(props.bucketName, key)
  emit('edit-key', key)
}

async function deleteKey(key: string) {
  closeContextMenu()
  if (confirm(`Delete key "${key}"?`)) {
    await kvStore.deleteKey(props.bucketName, key)
  }
}

async function purgeKey(key: string) {
  closeContextMenu()
  if (confirm(`Purge all revisions of key "${key}"?`)) {
    await kvStore.purgeKey(props.bucketName, key)
  }
}

function onKeySaved(key: string) {
  showCreateModal.value = false
  kvStore.fetchKeyDetail(props.bucketName, key)
}

// Close context menu on outside click
function onDocClick(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('[data-context-menu]')) {
    contextMenu.value = null
  }
}
</script>

<template>
  <div class="flex flex-col h-full border-r border-gray-200 dark:border-gray-800" @click.self="closeContextMenu" v-click-outside="onDocClick">
    <!-- Header -->
    <div class="flex items-center justify-between px-3 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0">
      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Keys</span>
      <div class="flex items-center gap-1">
        <span class="text-xs text-gray-400">{{ filteredKeys.length }} / {{ kvStore.keys.length }} keys</span>
        <button
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          title="Refresh"
          @click="kvStore.fetchKeys(bucketName)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
          </svg>
        </button>
        <button
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-emerald-600 dark:hover:text-emerald-400"
          title="Add key"
          @click="showCreateModal = true"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Search -->
    <div class="px-3 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0">
      <div class="relative">
        <svg class="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search"
          class="w-full pl-7 pr-3 py-1.5 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-emerald-500 focus:border-transparent"
        />
      </div>
    </div>

    <!-- Copy feedback toast -->
    <div
      v-if="copyFeedback"
      class="mx-3 mt-2 px-3 py-1.5 text-xs text-center rounded bg-gray-800 dark:bg-gray-200 text-white dark:text-gray-900 shrink-0"
    >
      {{ copyFeedback }}
    </div>

    <!-- Key list -->
    <div class="flex-1 overflow-y-auto">
      <div v-if="kvStore.keysLoading" class="px-3 py-4 text-xs text-gray-400 dark:text-gray-600 text-center">
        Loading...
      </div>
      <div v-else-if="filteredKeys.length === 0" class="px-3 py-4 text-xs text-gray-400 dark:text-gray-600 text-center italic">
        {{ searchQuery ? 'No matching keys' : 'No keys found' }}
      </div>
      <div v-else>
        <div
          v-for="key in filteredKeys"
          :key="key"
          class="group flex items-center gap-2 px-3 py-2 cursor-pointer border-b border-gray-100 dark:border-gray-800/50"
          :class="kvStore.selectedKey === key ? 'bg-emerald-50 dark:bg-emerald-950/40' : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'"
          @click="selectKey(key)"
          @contextmenu="openContextMenu($event, key)"
        >
          <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"/>
          </svg>
          <span class="text-sm text-gray-700 dark:text-gray-300 truncate flex-1">{{ key }}</span>
          <button
            class="opacity-0 group-hover:opacity-100 p-0.5 rounded hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 shrink-0"
            @click.stop="openContextMenu($event, key)"
          >
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 6a2 2 0 1 1 0-4 2 2 0 0 1 0 4Zm0 8a2 2 0 1 1 0-4 2 2 0 0 1 0 4Zm0 8a2 2 0 1 1 0-4 2 2 0 0 1 0 4Z"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Context menu (portal to body) -->
    <Teleport to="body">
      <div
        v-if="contextMenu"
        data-context-menu
        class="fixed z-50 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg py-1 min-w-36"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      >
        <button
          class="flex items-center gap-2.5 w-full px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
          @click="editKey(contextMenu.key)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125"/>
          </svg>
          Edit
        </button>
        <button
          class="flex items-center gap-2.5 w-full px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
          @click="copyKey(contextMenu.key)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z"/>
          </svg>
          Copy Key
        </button>
        <button
          class="flex items-center gap-2.5 w-full px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
          @click="copyValue(contextMenu.key)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184"/>
          </svg>
          Copy Value
        </button>
        <div class="my-1 border-t border-gray-100 dark:border-gray-800"></div>
        <button
          class="flex items-center gap-2.5 w-full px-3 py-1.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-950/30"
          @click="deleteKey(contextMenu.key)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
          </svg>
          Delete
        </button>
        <button
          class="flex items-center gap-2.5 w-full px-3 py-1.5 text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-950/30"
          @click="purgeKey(contextMenu.key)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"/>
          </svg>
          Purge
        </button>
      </div>
    </Teleport>

    <!-- Create key modal -->
    <KvKeyFormModal
      v-if="showCreateModal"
      :bucket-name="bucketName"
      @close="showCreateModal = false"
      @saved="onKeySaved"
    />
  </div>
</template>
