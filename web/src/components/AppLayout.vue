<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Toaster } from 'vue-sonner'
import { useUiStore } from '../stores/ui'
import { useConnectionsStore } from '../stores/connections'
import { useServerStore } from '../stores/server'
import { useKvStore } from '../stores/kv'
import { useStreamsStore } from '../stores/streams'
import { useObjectsStore } from '../stores/objects'
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'
import ConnectionPanel from '../features/connections/ConnectionPanel.vue'
import ManageConnectionsModal from '../features/connections/ManageConnectionsModal.vue'
import KvCreateBucketModal from '../features/kv/KvCreateBucketModal.vue'
import StreamCreateModal from '../features/streams/StreamCreateModal.vue'
import ObjectCreateStoreModal from '../features/objects/ObjectCreateStoreModal.vue'
import PublishModal from '../features/publish/PublishModal.vue'
import CommandPalette from './CommandPalette.vue'
import type { Connection } from '../lib/api'

const router = useRouter()

const uiStore = useUiStore()
const connStore = useConnectionsStore()
const serverStore = useServerStore()
const kvStore = useKvStore()
const streamsStore = useStreamsStore()
const objectsStore = useObjectsStore()

const isDark = computed(() => document.documentElement.classList.contains('dark'))

const isDragging = ref(false)

function onResizeMousedown(e: MouseEvent) {
  isDragging.value = true
  const startX = e.clientX
  const startWidth = uiStore.sidebarWidth

  function onMousemove(e: MouseEvent) {
    uiStore.setSidebarWidth(startWidth + e.clientX - startX)
  }

  function onMouseup() {
    isDragging.value = false
    document.body.classList.remove('select-none', 'cursor-col-resize')
    document.removeEventListener('mousemove', onMousemove)
    document.removeEventListener('mouseup', onMouseup)
  }

  document.body.classList.add('select-none', 'cursor-col-resize')
  document.addEventListener('mousemove', onMousemove)
  document.addEventListener('mouseup', onMouseup)
}

async function refreshAll() {
  await Promise.all([
    serverStore.fetchData(),
    streamsStore.fetchStreams(),
    kvStore.fetchBuckets(),
    objectsStore.fetchStores(),
  ])
}

const showConnectionModal = ref(false)
const showManageModal = ref(false)
const showKvCreateModal = ref(false)
const showStreamCreateModal = ref(false)
const showObjectStoreCreateModal = ref(false)
const showPublishModal = ref(false)
const editingConnection = ref<Connection | null>(null)

function openCreateKvBucket() {
  showKvCreateModal.value = true
}

function onKvBucketCreated(name: string) {
  showKvCreateModal.value = false
  router.push({ name: 'kv', params: { bucket: name } })
}

function openCreateStream() {
  showStreamCreateModal.value = true
}

function openCreateObjectStore() {
  showObjectStoreCreateModal.value = true
}

function openPublishMessage() {
  showPublishModal.value = true
}

function onStreamCreated(name: string) {
  showStreamCreateModal.value = false
  router.push({ name: 'streams', params: { stream: name } })
}

function onObjectStoreCreated(name: string) {
  showObjectStoreCreateModal.value = false
  router.push({ name: 'objects', params: { store: name } })
}

function openCreateConnection() {
  showManageModal.value = false
  editingConnection.value = null
  showConnectionModal.value = true
}

function openManageConnections() {
  showConnectionModal.value = false
  showManageModal.value = true
}

function editConnection(conn: Connection) {
  showManageModal.value = false
  editingConnection.value = conn
  showConnectionModal.value = true
}

function onConnectionSaved() {
  showConnectionModal.value = false
  editingConnection.value = null
}

// Fetch all data when connection changes (including switching between connections)
watch(() => connStore.status.connectionId, () => {
  if (connStore.status.connected) refreshAll()
  else { serverStore.clear(); streamsStore.clear(); kvStore.clear(); objectsStore.clear() }
})

onMounted(async () => {
  uiStore.init()
  await connStore.fetchConnections()
  await connStore.fetchStatus()
  if (connStore.status.connected) {
    refreshAll()
  }
})
</script>

<template>
  <div class="flex h-screen bg-white dark:bg-gray-950">
    <!-- Sidebar (only when connected) -->
    <AppSidebar
      v-if="connStore.status.connected && !uiStore.sidebarCollapsed"
      @create-kv-bucket="openCreateKvBucket"
      @create-stream="openCreateStream"
      @create-object-store="openCreateObjectStore"
      @publish-message="openPublishMessage"
    />
    <div
      v-if="connStore.status.connected && !uiStore.sidebarCollapsed"
      class="w-1 shrink-0 cursor-col-resize hover:bg-emerald-400 dark:hover:bg-emerald-600 active:bg-emerald-500"
      @mousedown.prevent="onResizeMousedown"
    />

    <!-- Main area -->
    <div class="flex-1 flex flex-col min-w-0">
      <AppHeader
        @create-connection="openCreateConnection"
        @manage-connections="openManageConnections"
      />
      <main class="flex-1 overflow-auto">
        <RouterView />
      </main>
    </div>

    <!-- Connection create/edit modal -->
    <ConnectionPanel
      v-if="showConnectionModal"
      :connection="editingConnection"
      @close="showConnectionModal = false"
      @saved="onConnectionSaved"
    />

    <!-- Manage connections modal -->
    <ManageConnectionsModal
      v-if="showManageModal"
      @close="showManageModal = false"
      @edit="editConnection"
      @create="openCreateConnection"
    />

    <!-- KV Bucket create modal (from sidebar +) -->
    <KvCreateBucketModal
      v-if="showKvCreateModal"
      @close="showKvCreateModal = false"
      @created="onKvBucketCreated"
    />

    <!-- Stream create modal (from sidebar +) -->
    <StreamCreateModal
      v-if="showStreamCreateModal"
      @close="showStreamCreateModal = false"
      @created="onStreamCreated"
    />

    <!-- Object Store create modal (from sidebar +) -->
    <ObjectCreateStoreModal
      v-if="showObjectStoreCreateModal"
      @close="showObjectStoreCreateModal = false"
      @created="onObjectStoreCreated"
    />

    <!-- Publish message modal -->
    <PublishModal
      v-if="showPublishModal"
      @close="showPublishModal = false"
      @published="showPublishModal = false"
    />

    <!-- Command palette (Cmd/Ctrl+K) -->
    <CommandPalette />
  </div>

  <Teleport to="body">
    <Toaster position="bottom-right" :theme="isDark ? 'dark' : 'light'" rich-colors />
  </Teleport>
</template>
