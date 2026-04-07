<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { useConnectionsStore } from '../stores/connections'
import { useServerStore } from '../stores/server'
import { useKvStore } from '../stores/kv'
import { useStreamsStore } from '../stores/streams'
import { useObjectsStore } from '../stores/objects'
import { useUiStore } from '../stores/ui'

const connStore = useConnectionsStore()
const serverStore = useServerStore()
const kvStore = useKvStore()
const streamsStore = useStreamsStore()
const objectsStore = useObjectsStore()
const uiStore = useUiStore()

const streamsOpen = ref(true)
const kvOpen = ref(true)
const objectsOpen = ref(true)

const searchOpen = ref(false)
const searchQuery = ref('')
const searchInputRef = ref<HTMLInputElement>()

function toggleSearch() {
  searchOpen.value = !searchOpen.value
  if (searchOpen.value) {
    nextTick(() => searchInputRef.value?.focus())
  } else {
    searchQuery.value = ''
  }
}

function clearSearch() {
  searchQuery.value = ''
  searchInputRef.value?.focus()
}

const filteredStreams = computed(() => {
  if (!searchQuery.value) return streamsStore.streams
  const q = searchQuery.value.toLowerCase()
  return streamsStore.streams.filter(s => s.config.name.toLowerCase().includes(q))
})

const filteredKvBuckets = computed(() => {
  if (!searchQuery.value) return kvStore.buckets
  const q = searchQuery.value.toLowerCase()
  return kvStore.buckets.filter(b => b.name.toLowerCase().includes(q))
})

const filteredObjectStores = computed(() => {
  if (!searchQuery.value) return objectsStore.stores
  const q = searchQuery.value.toLowerCase()
  return objectsStore.stores.filter(s => s.bucket.toLowerCase().includes(q))
})

const showCreateMenu = ref(false)
const createBtnRef = ref<HTMLElement>()
const menuStyle = ref({ top: '0px', left: '0px' })

function toggleCreateMenu() {
  showCreateMenu.value = !showCreateMenu.value
  if (showCreateMenu.value && createBtnRef.value) {
    const rect = createBtnRef.value.getBoundingClientRect()
    menuStyle.value = {
      top: `${rect.bottom + 4}px`,
      left: `${rect.left}px`,
    }
  }
}

function handleClickOutsideMenu(e: MouseEvent) {
  if (showCreateMenu.value && createBtnRef.value && !createBtnRef.value.contains(e.target as Node)) {
    showCreateMenu.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutsideMenu))
onUnmounted(() => document.removeEventListener('click', handleClickOutsideMenu))

const emit = defineEmits<{
  'create-kv-bucket': []
  'create-stream': []
  'create-object-store': []
  'publish-message': []
}>()

const navItems = [
  { name: 'Dashboard', path: '/', icon: 'dashboard' },
  { name: 'Account', path: '/account', icon: 'account' },
  { name: 'Live Tail', path: '/tail', icon: 'tail' },
]

async function refreshAll() {
  await Promise.all([
    serverStore.fetchData(),
    kvStore.fetchBuckets(),
    streamsStore.fetchStreams(),
    objectsStore.fetchStores(),
  ])
}

function createKvBucket() {
  showCreateMenu.value = false
  emit('create-kv-bucket')
}

function createStream() {
  showCreateMenu.value = false
  emit('create-stream')
}

function createObjectStore() {
  showCreateMenu.value = false
  emit('create-object-store')
}

function publishMessage() {
  showCreateMenu.value = false
  emit('publish-message')
}

onMounted(async () => {
  if (connStore.status.connected) {
    await Promise.all([
      kvStore.fetchBuckets(),
      streamsStore.fetchStreams(),
      objectsStore.fetchStores(),
    ])
  }
})
</script>

<template>
  <aside :style="{ width: uiStore.sidebarWidth + 'px' }" class="border-r border-gray-200 dark:border-gray-800 flex flex-col bg-white dark:bg-gray-950 shrink-0 overflow-y-auto">
    <!-- Toolbar -->
    <div class="flex items-center justify-end gap-1 px-2 py-1.5 border-b border-gray-200 dark:border-gray-800">
      <!-- Search -->
      <button
        class="p-1 rounded-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400"
        :class="searchOpen ? 'bg-gray-100 dark:bg-gray-800' : ''"
        title="Search"
        @click="toggleSearch"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"/>
        </svg>
      </button>

      <!-- Create dropdown -->
      <div>
        <button
          ref="createBtnRef"
          class="p-1 rounded-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400"
          title="Create"
          @click.stop="toggleCreateMenu"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
          </svg>
        </button>
        <Teleport to="body">
        <div
          v-if="showCreateMenu"
          class="fixed bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-sm shadow-lg z-[100] py-1 min-w-40"
          :style="menuStyle"
        >
          <div class="px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Create</div>
          <button
            class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
            @click="createStream"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"/>
            </svg>
            Stream
          </button>
          <button
            class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
            @click="createKvBucket"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"/>
            </svg>
            KV Bucket
          </button>
          <button
            class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
            @click="createObjectStore"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9"/>
            </svg>
            Object Store
          </button>
          <div class="my-1 border-t border-gray-200 dark:border-gray-700"></div>
          <div class="px-3 py-1 text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Publish</div>
          <button
            class="w-full flex items-center gap-2 px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
            @click="publishMessage"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5"/>
            </svg>
            Publish Message
          </button>
        </div>
        </Teleport>
      </div>

      <!-- Refresh -->
      <button
        class="p-1 rounded-sm hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400"
        title="Refresh all"
        @click="refreshAll"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182"/>
        </svg>
      </button>
    </div>

    <!-- Search input -->
    <div v-if="searchOpen" class="px-2 py-1.5 border-b border-gray-200 dark:border-gray-800">
      <div class="relative">
        <input
          ref="searchInputRef"
          v-model="searchQuery"
          type="text"
          placeholder="Filter..."
          class="w-full pl-2 pr-7 py-1 text-xs border border-gray-300 dark:border-gray-700 rounded bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-emerald-500 focus:border-transparent"
          @keydown.escape="toggleSearch"
        />
        <button
          v-if="searchQuery"
          class="absolute right-1.5 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          @click="clearSearch"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Main navigation -->
    <nav class="p-2 space-y-0.5">
      <RouterLink
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="flex items-center gap-2 px-3 py-1.5 text-xs rounded-none text-gray-500 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-900 font-mono"
        active-class="border-l-2 border-emerald-400 text-emerald-400 pl-[10px]"
      >
        <svg v-if="item.icon === 'dashboard'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25a2.25 2.25 0 0 1-2.25-2.25v-2.25Z"/>
        </svg>
        <svg v-else-if="item.icon === 'account'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"/>
        </svg>
        <svg v-else-if="item.icon === 'tail'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.288 15.038a5.25 5.25 0 0 1 7.424 0M5.106 11.856c3.807-3.808 9.98-3.808 13.788 0M1.924 8.674c5.565-5.565 14.587-5.565 20.152 0M12.53 18.22l-.53.53-.53-.53a.75.75 0 0 1 1.06 0Z"/>
        </svg>
        {{ item.name }}
      </RouterLink>
    </nav>

    <!-- Sections -->
    <div class="mt-2 flex-1">
      <!-- STREAMS -->
      <div class="mb-1">
        <button
          class="w-full px-4 py-1 flex items-center gap-1 text-[10px] font-semibold tracking-wider text-gray-400 dark:text-gray-600 uppercase hover:text-gray-600 dark:hover:text-gray-400"
          @click="streamsOpen = !streamsOpen"
        >
          <svg class="w-3 h-3 shrink-0" :class="streamsOpen ? '' : '-rotate-90'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5"/>
          </svg>
          <RouterLink
            :to="{ name: 'streams' }"
            class="hover:text-gray-600 dark:hover:text-gray-400"
            @click.stop
          >
            STREAMS ({{ searchQuery ? filteredStreams.length + '/' + streamsStore.streams.length : streamsStore.streams.length }})
          </RouterLink>
        </button>
        <template v-if="streamsOpen">
          <template v-if="filteredStreams.length > 0">
            <RouterLink
              v-for="s in filteredStreams"
              :key="s.config.name"
              :to="{ name: 'streams', params: { stream: s.config.name } }"
              class="flex items-center gap-2 px-4 py-1 text-xs text-gray-500 hover:text-gray-900 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-900 font-mono"
              active-class="border-l-2 border-emerald-400 text-emerald-400 pl-[14px]"
            >
              <svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"/>
              </svg>
              <span class="truncate">{{ s.config.name }}</span>
            </RouterLink>
          </template>
          <div v-else class="px-4 py-1 text-xs text-gray-400 dark:text-gray-600 italic">
            Nothing found
          </div>
        </template>
      </div>

      <!-- KV BUCKETS -->
      <div class="mb-1">
        <button
          class="w-full px-4 py-1 flex items-center gap-1 text-[10px] font-semibold tracking-wider text-gray-400 dark:text-gray-600 uppercase hover:text-gray-600 dark:hover:text-gray-400"
          @click="kvOpen = !kvOpen"
        >
          <svg class="w-3 h-3 shrink-0" :class="kvOpen ? '' : '-rotate-90'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5"/>
          </svg>
          <RouterLink
            :to="{ name: 'kv' }"
            class="hover:text-gray-600 dark:hover:text-gray-400"
            @click.stop
          >
            KV BUCKETS ({{ searchQuery ? filteredKvBuckets.length + '/' + kvStore.buckets.length : kvStore.buckets.length }})
          </RouterLink>
        </button>
        <template v-if="kvOpen">
          <template v-if="filteredKvBuckets.length > 0">
            <RouterLink
              v-for="bucket in filteredKvBuckets"
              :key="bucket.name"
              :to="{ name: 'kv', params: { bucket: bucket.name } }"
              class="flex items-center gap-2 px-4 py-1 text-xs text-gray-500 hover:text-gray-900 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-900 font-mono"
              active-class="border-l-2 border-emerald-400 text-emerald-400 pl-[14px]"
            >
              <svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"/>
              </svg>
              <span class="truncate">{{ bucket.name }}</span>
            </RouterLink>
          </template>
          <div v-else class="px-4 py-1 text-xs text-gray-400 dark:text-gray-600 italic">
            Nothing found
          </div>
        </template>
      </div>

      <!-- OBJECT BUCKETS -->
      <div class="mb-1">
        <button
          class="w-full px-4 py-1 flex items-center gap-1 text-[10px] font-semibold tracking-wider text-gray-400 dark:text-gray-600 uppercase hover:text-gray-600 dark:hover:text-gray-400"
          @click="objectsOpen = !objectsOpen"
        >
          <svg class="w-3 h-3 shrink-0" :class="objectsOpen ? '' : '-rotate-90'" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5"/>
          </svg>
          <RouterLink
            :to="{ name: 'objects' }"
            class="hover:text-gray-600 dark:hover:text-gray-400"
            @click.stop
          >
            OBJECT BUCKETS ({{ searchQuery ? filteredObjectStores.length + '/' + objectsStore.stores.length : objectsStore.stores.length }})
          </RouterLink>
        </button>
        <template v-if="objectsOpen">
          <template v-if="filteredObjectStores.length > 0">
            <RouterLink
              v-for="store in filteredObjectStores"
              :key="store.bucket"
              :to="{ name: 'objects', params: { store: store.bucket } }"
              class="flex items-center gap-2 px-4 py-1 text-xs text-gray-500 hover:text-gray-900 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-900 font-mono"
              active-class="border-l-2 border-emerald-400 text-emerald-400 pl-[14px]"
            >
              <svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M21 7.5l-9-5.25L3 7.5m18 0l-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9"/>
              </svg>
              <span class="truncate">{{ store.bucket }}</span>
            </RouterLink>
          </template>
          <div v-else class="px-4 py-1 text-xs text-gray-400 dark:text-gray-600 italic">
            Nothing found
          </div>
        </template>
      </div>
    </div>

    <!-- Connection status -->
    <div class="p-3 border-t border-gray-200 dark:border-gray-800">
      <div class="flex items-center gap-2 px-2 py-1 text-xs text-gray-500">
        <span class="w-2 h-2 rounded-full shrink-0 bg-emerald-400"></span>
        Connected
      </div>
    </div>
  </aside>
</template>
