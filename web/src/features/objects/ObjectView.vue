<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useObjectsStore } from '../../stores/objects'
import { useConnectionsStore } from '../../stores/connections'
import ObjectList from './ObjectList.vue'
import ObjectConfigView from './ObjectConfigView.vue'
import ObjectCreateStoreModal from './ObjectCreateStoreModal.vue'

type Tab = 'objects' | 'config'

const route = useRoute()
const router = useRouter()
const objectsStore = useObjectsStore()
const connStore = useConnectionsStore()

const activeTab = ref<Tab>('objects')
const showCreateModal = ref(false)
const confirmDeleteStore = ref(false)

const selectedStoreName = computed(() => route.params.store as string | undefined)
const hasStore = computed(() => !!selectedStoreName.value)

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

async function onStoreCreated(name: string) {
  showCreateModal.value = false
  await router.push({ name: 'objects', params: { store: name } })
}

async function handleDeleteStore() {
  if (!selectedStoreName.value) return
  if (!confirmDeleteStore.value) {
    confirmDeleteStore.value = true
    return
  }
  await objectsStore.deleteStore(selectedStoreName.value)
  confirmDeleteStore.value = false
  router.push({ name: 'objects' })
}

watch(selectedStoreName, async (name) => {
  activeTab.value = 'objects'
  confirmDeleteStore.value = false
  if (name && connStore.status.connected) {
    await objectsStore.selectStore(name)
  } else {
    objectsStore.selectedStore = null
    objectsStore.objects = []
  }
}, { immediate: true })

onMounted(async () => {
  if (connStore.status.connected) {
    await objectsStore.fetchStores()
  }
  if (route.query.create) {
    showCreateModal.value = true
    router.replace({ query: {} })
  }
})
</script>

<template>
  <div class="flex flex-col h-full">

    <template v-if="!hasStore">
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
        <h1 class="text-base font-semibold text-gray-900 dark:text-gray-100">Object Stores</h1>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
          @click="showCreateModal = true"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
          </svg>
          Create Object Store
        </button>
      </div>

      <div class="flex-1 overflow-auto p-6">
        <div v-if="objectsStore.storesLoading" class="text-sm text-gray-400 dark:text-gray-600">Loading stores...</div>
        <div v-else-if="objectsStore.storesError" class="text-sm text-red-500">{{ objectsStore.storesError }}</div>
        <div v-else-if="objectsStore.stores.length === 0" class="flex flex-col items-center justify-center py-20 gap-4">
          <svg class="w-12 h-12 text-gray-300 dark:text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125"/>
          </svg>
          <p class="text-sm text-gray-500 dark:text-gray-400">No object stores found</p>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
            @click="showCreateModal = true"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
            </svg>
            Create your first store
          </button>
        </div>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <RouterLink
            v-for="store in objectsStore.stores"
            :key="store.bucket"
            :to="{ name: 'objects', params: { store: store.bucket } }"
            class="flex flex-col gap-2 p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-emerald-400 dark:hover:border-emerald-600 bg-white dark:bg-gray-900"
          >
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125"/>
              </svg>
              <span class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">{{ store.bucket }}</span>
            </div>
            <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
              <span>{{ formatBytes(store.size) }}</span>
              <span class="capitalize">{{ store.storage }}</span>
              <span>R{{ store.replicas }}</span>
            </div>
          </RouterLink>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="border-b border-gray-200 dark:border-gray-800 shrink-0">
        <div class="flex items-center justify-between px-6 pt-3 pb-0">
          <div class="flex items-center gap-2">
            <RouterLink :to="{ name: 'objects' }" class="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
              Object Stores
            </RouterLink>
            <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
            </svg>
            <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">{{ selectedStoreName }}</span>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              title="Refresh"
              @click="objectsStore.selectStore(selectedStoreName!)"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
              </svg>
            </button>
            <button
              class="flex items-center gap-1 px-3 py-1 text-xs rounded-md"
              :class="confirmDeleteStore
                ? 'bg-red-600 text-white hover:bg-red-700'
                : 'text-red-500 hover:bg-red-50 dark:hover:bg-red-950'"
              @click="handleDeleteStore"
            >
              {{ confirmDeleteStore ? 'Click to confirm' : 'Delete Store' }}
            </button>
          </div>
        </div>

        <div v-if="objectsStore.selectedStore" class="flex items-center gap-4 px-6 py-1.5 text-xs text-gray-500 dark:text-gray-400">
          <span>Size <strong class="text-gray-700 dark:text-gray-300">{{ formatBytes(objectsStore.selectedStore.size) }}</strong></span>
          <span>Storage <strong class="text-gray-700 dark:text-gray-300 capitalize">{{ objectsStore.selectedStore.storage }}</strong></span>
          <span>Replicas <strong class="text-gray-700 dark:text-gray-300">R{{ objectsStore.selectedStore.replicas }}</strong></span>
          <span v-if="objectsStore.selectedStore.sealed" class="text-amber-600 dark:text-amber-400 font-medium">Sealed</span>
        </div>

        <div class="flex items-center gap-0 px-6 pt-1">
          <button
            v-for="tab in (['objects', 'config'] as Tab[])"
            :key="tab"
            class="px-4 py-2 text-sm capitalize border-b-2"
            :class="activeTab === tab
              ? 'border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 font-medium'
              : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
            @click="activeTab = tab"
          >
            {{ tab }}
          </button>
        </div>
      </div>

      <div class="flex-1 overflow-hidden">
        <ObjectList v-if="activeTab === 'objects'" :store-name="selectedStoreName!" class="h-full" />
        <ObjectConfigView v-else-if="activeTab === 'config'" />
      </div>
    </template>

    <ObjectCreateStoreModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="onStoreCreated"
    />
  </div>
</template>
