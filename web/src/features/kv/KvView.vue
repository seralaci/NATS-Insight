<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useKvStore } from '../../stores/kv'
import { useConnectionsStore } from '../../stores/connections'
import KvKeyList from './KvKeyList.vue'
import KvKeyDetail from './KvKeyDetail.vue'
import KvCreateBucketModal from './KvCreateBucketModal.vue'
import KvConfigView from './KvConfigView.vue'
import KvWatchPanel from './KvWatchPanel.vue'

type Tab = 'content' | 'config' | 'watch'

const route = useRoute()
const router = useRouter()
const kvStore = useKvStore()
const connStore = useConnectionsStore()

const activeTab = ref<Tab>('content')
const showCreateModal = ref(false)
const confirmDeleteBucket = ref(false)
const keyDetailRef = ref<InstanceType<typeof KvKeyDetail> | null>(null)

function handleEditKey() {
  // Wait a tick for keyDetail to load, then open edit modal
  setTimeout(() => keyDetailRef.value?.openEdit(), 100)
}

const selectedBucketName = computed(() => route.params.bucket as string | undefined)
const hasBucket = computed(() => !!selectedBucketName.value)

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

async function onBucketCreated(name: string) {
  showCreateModal.value = false
  await router.push({ name: 'kv', params: { bucket: name } })
}

async function handleDeleteBucket() {
  if (!selectedBucketName.value) return
  if (!confirmDeleteBucket.value) {
    confirmDeleteBucket.value = true
    return
  }
  await kvStore.deleteBucket(selectedBucketName.value)
  confirmDeleteBucket.value = false
  router.push({ name: 'kv' })
}

// Load bucket when route param changes
watch(selectedBucketName, async (name) => {
  activeTab.value = 'content'
  confirmDeleteBucket.value = false
  if (name && connStore.status.connected) {
    await kvStore.selectBucket(name)
  } else {
    kvStore.selectedBucket = null
    kvStore.keys = []
    kvStore.selectedKey = null
    kvStore.keyDetail = null
    kvStore.keyHistory = []
  }
}, { immediate: true })

onMounted(async () => {
  if (connStore.status.connected) {
    await kvStore.fetchBuckets()
  }
  // Open create modal if ?create=1 query param is present
  if (route.query.create) {
    showCreateModal.value = true
    router.replace({ query: {} })
  }
})
</script>

<template>
  <div class="flex flex-col h-full">

    <!-- No bucket selected — bucket list -->
    <template v-if="!hasBucket">
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
        <h1 class="text-base font-semibold text-gray-900 dark:text-gray-100">KV Buckets</h1>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
          @click="showCreateModal = true"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
          </svg>
          Create KV Bucket
        </button>
      </div>

      <div class="flex-1 overflow-auto p-6">
        <div v-if="kvStore.bucketsLoading" class="text-sm text-gray-400 dark:text-gray-600">Loading buckets...</div>
        <div v-else-if="kvStore.bucketsError" class="text-sm text-red-500">{{ kvStore.bucketsError }}</div>
        <div v-else-if="kvStore.buckets.length === 0" class="flex flex-col items-center justify-center py-20 gap-4">
          <svg class="w-12 h-12 text-gray-300 dark:text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"/>
          </svg>
          <p class="text-sm text-gray-500 dark:text-gray-400">No KV buckets found</p>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
            @click="showCreateModal = true"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
            </svg>
            Create your first bucket
          </button>
        </div>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <RouterLink
            v-for="bucket in kvStore.buckets"
            :key="bucket.name"
            :to="{ name: 'kv', params: { bucket: bucket.name } }"
            class="flex flex-col gap-2 p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-emerald-400 dark:hover:border-emerald-600 bg-white dark:bg-gray-900"
          >
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"/>
              </svg>
              <span class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">{{ bucket.name }}</span>
            </div>
            <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
              <span>{{ bucket.values }} values</span>
              <span>{{ formatBytes(bucket.bytes) }}</span>
              <span>R{{ bucket.replicas }}</span>
            </div>
          </RouterLink>
        </div>
      </div>
    </template>

    <!-- Bucket selected -->
    <template v-else>
      <!-- Bucket header with tabs -->
      <div class="border-b border-gray-200 dark:border-gray-800 shrink-0">
        <!-- Bucket name + actions -->
        <div class="flex items-center justify-between px-6 pt-3 pb-0">
          <div class="flex items-center gap-2">
            <RouterLink :to="{ name: 'kv' }" class="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
              KV Buckets
            </RouterLink>
            <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
            </svg>
            <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">{{ selectedBucketName }}</span>
          </div>
          <div class="flex items-center gap-2">
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              title="Refresh"
              @click="kvStore.selectBucket(selectedBucketName!)"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
              </svg>
            </button>
            <button
              class="flex items-center gap-1 px-3 py-1 text-xs rounded-md"
              :class="confirmDeleteBucket
                ? 'bg-red-600 text-white hover:bg-red-700'
                : 'text-red-500 hover:bg-red-50 dark:hover:bg-red-950'"
              @click="handleDeleteBucket"
            >
              {{ confirmDeleteBucket ? 'Click to confirm' : 'Delete Bucket' }}
            </button>
          </div>
        </div>

        <!-- Stats bar -->
        <div v-if="kvStore.selectedBucket" class="flex items-center gap-4 px-6 py-1.5 text-xs text-gray-500 dark:text-gray-400">
          <span>Values <strong class="text-gray-700 dark:text-gray-300">{{ kvStore.selectedBucket.values }}</strong></span>
          <span>Bucket Size <strong class="text-gray-700 dark:text-gray-300">{{ formatBytes(kvStore.selectedBucket.bytes) }}</strong></span>
          <span>History Values <strong class="text-gray-700 dark:text-gray-300">{{ kvStore.selectedBucket.history }}</strong></span>
          <span>Replicas <strong class="text-gray-700 dark:text-gray-300">R{{ kvStore.selectedBucket.replicas }}</strong></span>
          <span>Storage <strong class="text-gray-700 dark:text-gray-300 capitalize">{{ kvStore.selectedBucket.storage }}</strong></span>
        </div>

        <!-- Tab bar -->
        <div class="flex items-center gap-0 px-6 pt-1">
          <button
            v-for="tab in (['content', 'config', 'watch'] as Tab[])"
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

      <!-- Tab content -->
      <div class="flex-1 overflow-hidden">
        <!-- Content tab -->
        <div v-if="activeTab === 'content'" class="flex h-full">
          <div class="w-72 shrink-0">
            <KvKeyList :bucket-name="selectedBucketName!" @edit-key="handleEditKey" />
          </div>
          <div class="flex-1 min-w-0">
            <KvKeyDetail ref="keyDetailRef" :bucket-name="selectedBucketName!" />
          </div>
        </div>

        <!-- Config tab -->
        <KvConfigView v-else-if="activeTab === 'config'" />

        <!-- Watch tab -->
        <KvWatchPanel v-else-if="activeTab === 'watch'" :bucket="selectedBucketName!" class="h-full" />
      </div>
    </template>

    <!-- Create bucket modal -->
    <KvCreateBucketModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="onBucketCreated"
    />
  </div>
</template>
