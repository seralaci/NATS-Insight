<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useStreamsStore } from '../stores/streams'
import { useKvStore } from '../stores/kv'
import { useObjectsStore } from '../stores/objects'

const router = useRouter()
const streamsStore = useStreamsStore()
const kvStore = useKvStore()
const objectsStore = useObjectsStore()

const show = ref(false)
const query = ref('')
const selectedIndex = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)
const listRef = ref<HTMLDivElement | null>(null)

interface CommandItem {
  id: string
  label: string
  category: string
  action: () => void
}

const staticCommands: CommandItem[] = [
  { id: 'nav-dashboard', label: 'Go to Dashboard', category: 'Navigation', action: () => router.push('/') },
  { id: 'nav-streams', label: 'Go to Streams', category: 'Navigation', action: () => router.push('/streams') },
  { id: 'nav-kv', label: 'Go to Key-Value', category: 'Navigation', action: () => router.push('/kv') },
  { id: 'nav-objects', label: 'Go to Object Store', category: 'Navigation', action: () => router.push('/objects') },
  { id: 'nav-tail', label: 'Go to Live Tail', category: 'Navigation', action: () => router.push('/tail') },
  { id: 'nav-account', label: 'Go to Account', category: 'Navigation', action: () => router.push('/account') },
]

const filteredItems = computed(() => {
  const q = query.value.toLowerCase().trim()

  const dynamicItems: CommandItem[] = []

  for (const s of streamsStore.streams) {
    dynamicItems.push({
      id: `stream-${s.config.name}`,
      label: s.config.name,
      category: 'Streams',
      action: () => router.push({ name: 'streams', params: { stream: s.config.name } }),
    })
  }

  for (const b of kvStore.buckets) {
    dynamicItems.push({
      id: `kv-${b.name}`,
      label: b.name,
      category: 'KV Buckets',
      action: () => router.push({ name: 'kv', params: { bucket: b.name } }),
    })
  }

  for (const o of objectsStore.stores) {
    dynamicItems.push({
      id: `obj-${o.bucket}`,
      label: o.bucket,
      category: 'Object Stores',
      action: () => router.push({ name: 'objects', params: { store: o.bucket } }),
    })
  }

  const allItems = [...staticCommands, ...dynamicItems]

  if (!q) return allItems.slice(0, 20)
  return allItems.filter(item =>
    item.label.toLowerCase().includes(q) || item.category.toLowerCase().includes(q)
  ).slice(0, 20)
})

watch(filteredItems, () => {
  selectedIndex.value = 0
})

watch(selectedIndex, (i) => {
  nextTick(() => {
    if (!listRef.value) return
    const el = listRef.value.children[i] as HTMLElement | undefined
    el?.scrollIntoView({ block: 'nearest' })
  })
})

function toggle() {
  show.value = !show.value
  if (show.value) {
    query.value = ''
    selectedIndex.value = 0
    nextTick(() => inputRef.value?.focus())
  }
}

function close() {
  show.value = false
}

function execute(item: CommandItem) {
  item.action()
  close()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, filteredItems.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    const item = filteredItems.value[selectedIndex.value]
    if (item) execute(item)
  } else if (e.key === 'Escape') {
    close()
  }
}

function onGlobalKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    toggle()
  }
}

onMounted(() => {
  document.addEventListener('keydown', onGlobalKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onGlobalKeydown)
})
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[70] flex items-start justify-center pt-[20vh]">
      <div class="absolute inset-0 bg-black/50" @click="close"></div>
      <div class="relative w-full max-w-lg bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4 overflow-hidden">
        <!-- Search input -->
        <div class="flex items-center gap-3 px-4 py-3 border-b border-gray-200 dark:border-gray-800">
          <svg class="w-4 h-4 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"/>
          </svg>
          <input
            ref="inputRef"
            v-model="query"
            type="text"
            placeholder="Search streams, buckets, or type a command..."
            class="flex-1 bg-transparent text-sm text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-600 outline-none"
            @keydown="handleKeydown"
          />
          <kbd class="hidden sm:inline-flex items-center px-1.5 py-0.5 text-[10px] font-medium text-gray-400 dark:text-gray-600 bg-gray-100 dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700">ESC</kbd>
        </div>

        <!-- Results -->
        <div ref="listRef" class="max-h-72 overflow-y-auto py-2">
          <div v-if="filteredItems.length === 0" class="px-4 py-8 text-center text-sm text-gray-400 dark:text-gray-600">
            No results found
          </div>
          <template v-else>
            <div
              v-for="(item, i) in filteredItems"
              :key="item.id"
              class="flex items-center gap-3 px-4 py-2 cursor-pointer text-sm"
              :class="i === selectedIndex ? 'bg-emerald-50 dark:bg-emerald-950/30 text-emerald-700 dark:text-emerald-300' : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-900'"
              @click="execute(item)"
              @mouseenter="selectedIndex = i"
            >
              <span class="text-xs font-medium text-gray-400 dark:text-gray-600 w-24 shrink-0 truncate">{{ item.category }}</span>
              <span class="truncate font-mono text-xs">{{ item.label }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>
  </Teleport>
</template>
