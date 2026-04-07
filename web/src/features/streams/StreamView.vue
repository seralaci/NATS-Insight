<script setup lang="ts">
import { ref, watch, onMounted, computed, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useStreamsStore } from '../../stores/streams'
import { useConnectionsStore } from '../../stores/connections'
import StreamMessages from './StreamMessages.vue'
import StreamConfigView from './StreamConfigView.vue'
import StreamConsumers from './StreamConsumers.vue'
import StreamCreateModal from './StreamCreateModal.vue'
import StreamMessageDetail from './StreamMessageDetail.vue'
import StreamMessageInspector from './StreamMessageInspector.vue'
import StreamConsumerDetail from './StreamConsumerDetail.vue'
import type { StreamMessage, ConsumerInfo } from '../../lib/api'

type Tab = 'messages' | 'config' | 'consumers'

const route = useRoute()
const router = useRouter()
const streamsStore = useStreamsStore()
const connStore = useConnectionsStore()

const activeTab = ref<Tab>('messages')
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showMirrorModal = ref(false)
const showDuplicateModal = ref(false)
const confirmDelete = ref(false)
const showPurgeModal = ref(false)
const purgeSubject = ref('')
const purgeError = ref('')

// Three dots menu
const showDotsMenu = ref(false)
const dotsMenuRef = ref<HTMLElement | null>(null)
const dotsBtnRef = ref<HTMLElement | null>(null)

function toggleDotsMenu() {
  showDotsMenu.value = !showDotsMenu.value
  confirmDelete.value = false
}

function onDotsDocClick(e: MouseEvent) {
  if (!showDotsMenu.value) return
  const target = e.target as Node
  if (dotsMenuRef.value?.contains(target) || dotsBtnRef.value?.contains(target)) return
  showDotsMenu.value = false
}

onUnmounted(() => {
  document.removeEventListener('mousedown', onDotsDocClick)
})

// Sub-navigation state
const showInspector = ref(false)
const selectedMessage = ref<StreamMessage | null>(null)
const selectedConsumerDetail = ref<ConsumerInfo | null>(null)

const selectedStreamName = computed(() => route.params.stream as string | undefined)
const hasStream = computed(() => !!selectedStreamName.value)

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

async function onStreamCreated(name: string) {
  showCreateModal.value = false
  showMirrorModal.value = false
  showDuplicateModal.value = false
  await router.push({ name: 'streams', params: { stream: name } })
}

async function onStreamUpdated(name: string) {
  showEditModal.value = false
  await streamsStore.selectStream(name)
}

function handleSelectMessage(msg: StreamMessage) {
  selectedMessage.value = msg
}

function handleBackFromMessage() {
  selectedMessage.value = null
}

function handleSelectConsumer(consumer: ConsumerInfo) {
  selectedConsumerDetail.value = consumer
}

function handleBackFromConsumer() {
  selectedConsumerDetail.value = null
}

function handleConsumerRefresh() {
  if (selectedConsumerDetail.value) {
    const updated = streamsStore.consumers.find(c => c.name === selectedConsumerDetail.value?.name)
    if (updated) selectedConsumerDetail.value = updated
  }
}

function switchTab(tab: Tab) {
  activeTab.value = tab
  selectedMessage.value = null
  selectedConsumerDetail.value = null
  showInspector.value = false
}

async function handleDeleteStream() {
  if (!selectedStreamName.value) return
  if (!confirmDelete.value) {
    confirmDelete.value = true
    return
  }
  await streamsStore.deleteStream(selectedStreamName.value)
  confirmDelete.value = false
  showDotsMenu.value = false
  router.push({ name: 'streams' })
}

async function handlePurge() {
  if (!selectedStreamName.value) return
  purgeError.value = ''
  try {
    const opts: any = {}
    if (purgeSubject.value) opts.subject = purgeSubject.value
    await streamsStore.purgeStream(selectedStreamName.value, opts)
    showPurgeModal.value = false
    purgeSubject.value = ''
    if (activeTab.value === 'messages') {
      await streamsStore.fetchMessages(selectedStreamName.value, { limit: 50 })
    }
  } catch (e: any) {
    purgeError.value = e.message
  }
}

watch(selectedStreamName, async (name) => {
  activeTab.value = 'messages'
  confirmDelete.value = false
  selectedMessage.value = null
  selectedConsumerDetail.value = null
  showInspector.value = false
  if (name && connStore.status.connected) {
    await streamsStore.selectStream(name)
  } else {
    streamsStore.selectedStream = null
    streamsStore.messages = []
    streamsStore.consumers = []
  }
}, { immediate: true })

onMounted(async () => {
  document.addEventListener('mousedown', onDotsDocClick)
  if (connStore.status.connected) {
    await streamsStore.fetchStreams()
  }
  if (route.query.create) {
    showCreateModal.value = true
    router.replace({ query: {} })
  }
})
</script>

<template>
  <div class="flex flex-col h-full">

    <!-- No stream selected — stream grid -->
    <template v-if="!hasStream">
      <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
        <h1 class="text-base font-semibold text-gray-900 dark:text-gray-100">Streams</h1>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
          @click="showCreateModal = true"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
          </svg>
          Create Stream
        </button>
      </div>

      <div class="flex-1 overflow-auto p-6">
        <div v-if="streamsStore.streamsLoading" class="text-sm text-gray-400 dark:text-gray-600">Loading streams...</div>
        <div v-else-if="streamsStore.streamsError" class="text-sm text-red-500">{{ streamsStore.streamsError }}</div>
        <div v-else-if="streamsStore.streams.length === 0" class="flex flex-col items-center justify-center py-20 gap-4">
          <svg class="w-12 h-12 text-gray-300 dark:text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"/>
          </svg>
          <p class="text-sm text-gray-500 dark:text-gray-400">No streams found</p>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
            @click="showCreateModal = true"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
            </svg>
            Create your first stream
          </button>
        </div>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <RouterLink
            v-for="s in streamsStore.streams"
            :key="s.config.name"
            :to="{ name: 'streams', params: { stream: s.config.name } }"
            class="flex flex-col gap-3 p-4 border border-gray-200 dark:border-gray-800 rounded-lg hover:border-emerald-400 dark:hover:border-emerald-600 bg-white dark:bg-gray-900"
          >
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12h16.5m-16.5 3.75h16.5M3.75 19.5h16.5M5.625 4.5h12.75a1.875 1.875 0 0 1 0 3.75H5.625a1.875 1.875 0 0 1 0-3.75Z"/>
              </svg>
              <span class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">{{ s.config.name }}</span>
            </div>
            <div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
              <span>Messages <strong class="text-gray-700 dark:text-gray-300">{{ s.state.messages }}</strong></span>
              <span>Bytes <strong class="text-gray-700 dark:text-gray-300">{{ formatBytes(s.state.bytes) }}</strong></span>
              <span>Consumers <strong class="text-gray-700 dark:text-gray-300">{{ s.state.consumer_count }}</strong></span>
              <span>Storage <strong class="text-gray-700 dark:text-gray-300 capitalize">{{ s.config.storage }}</strong></span>
              <span>Retention <strong class="text-gray-700 dark:text-gray-300 capitalize">{{ s.config.retention }}</strong></span>
              <span>Replicas <strong class="text-gray-700 dark:text-gray-300">R{{ s.config.replicas }}</strong></span>
            </div>
          </RouterLink>
        </div>
      </div>
    </template>

    <!-- Stream selected -->
    <template v-else>
      <!-- Header (hidden when viewing detail views) -->
      <div v-if="!selectedMessage && !selectedConsumerDetail" class="border-b border-gray-200 dark:border-gray-800 shrink-0">
        <!-- Tab bar row: breadcrumb tab + tabs + three dots -->
        <div class="flex items-end gap-0 px-4 pt-0">

          <!-- Breadcrumb "tab" showing stream name -->
          <div class="flex items-center gap-1.5 px-3 py-2 mr-2">
            <RouterLink :to="{ name: 'streams' }" class="text-xs text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
              Streams
            </RouterLink>
            <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
            </svg>
            <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">{{ selectedStreamName }}</span>
          </div>

          <!-- Tab buttons -->
          <button
            v-for="tab in (['messages', 'config', 'consumers'] as Tab[])"
            :key="tab"
            class="px-4 py-2 text-sm capitalize border-b-2"
            :class="activeTab === tab
              ? 'border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 font-medium'
              : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
            @click="switchTab(tab)"
          >
            {{ tab === 'config' ? 'Config & State' : tab === 'consumers' ? 'Consumers' : 'Messages' }}
          </button>

          <!-- Three dots menu -->
          <div class="relative flex items-center ml-1 mb-0.5 self-center">
            <button
              ref="dotsBtnRef"
              class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-800 text-gray-500"
              title="Manage Stream"
              @click="toggleDotsMenu"
            >
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                <circle cx="5" cy="12" r="1.5"/>
                <circle cx="12" cy="12" r="1.5"/>
                <circle cx="19" cy="12" r="1.5"/>
              </svg>
            </button>

            <!-- Dropdown menu -->
            <div
              v-if="showDotsMenu"
              ref="dotsMenuRef"
              class="absolute left-0 top-full mt-1 w-52 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 py-1"
            >
              <p class="px-3 pt-1 pb-1 text-xs font-medium text-gray-400 dark:text-gray-500 uppercase tracking-wider">Manage Stream</p>

              <!-- Edit Stream -->
              <button
                class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                @click="showEditModal = true; showDotsMenu = false"
              >
                <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125"/>
                </svg>
                Edit Stream
              </button>

              <!-- Create Mirror -->
              <button
                class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                @click="showMirrorModal = true; showDotsMenu = false"
              >
                <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5"/>
                </svg>
                Create Mirror
              </button>

              <!-- Duplicate Stream -->
              <button
                class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                @click="showDuplicateModal = true; showDotsMenu = false"
              >
                <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 0 1-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 0 1 1.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.243-8.161-7.5-8.876a9.06 9.06 0 0 0-1.5-.124H9.375c-.621 0-1.125.504-1.125 1.125v3.5m7.5 10.375H9.375a1.125 1.125 0 0 1-1.125-1.125v-9.25m12 6.625v-1.875a3.375 3.375 0 0 0-3.375-3.375h-1.5a1.125 1.125 0 0 1-1.125-1.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H9.75"/>
                </svg>
                Duplicate Stream
              </button>

              <!-- Purge Messages -->
              <button
                class="w-full flex items-center gap-2 px-3 py-2 text-xs text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                @click="showPurgeModal = true; showDotsMenu = false"
              >
                <svg class="w-3.5 h-3.5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
                </svg>
                Purge Messages
              </button>

              <div class="border-t border-gray-100 dark:border-gray-800 my-1"></div>

              <!-- Delete Stream -->
              <button
                class="w-full flex items-center gap-2 px-3 py-2 text-xs hover:bg-red-50 dark:hover:bg-red-950/40"
                :class="confirmDelete ? 'text-red-700 dark:text-red-400 font-medium' : 'text-red-500 dark:text-red-400'"
                @click="handleDeleteStream"
              >
                <svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
                </svg>
                {{ confirmDelete ? 'Click to confirm delete' : 'Delete Stream' }}
              </button>
            </div>
          </div>

          <!-- Spacer + Refresh -->
          <div class="flex-1"></div>
          <div class="flex items-center mb-1">
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              title="Refresh"
              @click="streamsStore.selectStream(selectedStreamName!)"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
              </svg>
            </button>
          </div>
        </div>
      </div>

      <!-- Tab content -->
      <div class="flex-1 overflow-hidden">
        <!-- Messages tab: list or detail or inspector -->
        <template v-if="activeTab === 'messages'">
          <StreamMessageInspector
            v-if="showInspector"
            :stream-name="selectedStreamName!"
            class="h-full"
            @back="showInspector = false"
          />
          <StreamMessageDetail
            v-else-if="selectedMessage"
            :message="selectedMessage"
            :stream-name="selectedStreamName!"
            class="h-full"
            @back="handleBackFromMessage"
          />
          <StreamMessages
            v-else
            :stream-name="selectedStreamName!"
            class="h-full"
            @select-message="handleSelectMessage"
            @open-inspector="showInspector = true"
          />
        </template>

        <!-- Config tab -->
        <StreamConfigView v-else-if="activeTab === 'config'" />

        <!-- Consumers tab: list or detail -->
        <template v-else-if="activeTab === 'consumers'">
          <StreamConsumerDetail
            v-if="selectedConsumerDetail"
            :consumer="selectedConsumerDetail"
            :stream-name="selectedStreamName!"
            class="h-full"
            @back="handleBackFromConsumer"
            @refresh="handleConsumerRefresh"
          />
          <StreamConsumers
            v-else
            :stream-name="selectedStreamName!"
            class="h-full"
            @select-consumer="handleSelectConsumer"
          />
        </template>
      </div>
    </template>

    <!-- Create modal -->
    <StreamCreateModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="onStreamCreated"
    />

    <!-- Edit modal -->
    <StreamCreateModal
      v-if="showEditModal && streamsStore.selectedStream"
      :edit-stream="streamsStore.selectedStream"
      @close="showEditModal = false"
      @updated="onStreamUpdated"
    />

    <!-- Mirror modal -->
    <StreamCreateModal
      v-if="showMirrorModal && streamsStore.selectedStream"
      :mirror-stream="streamsStore.selectedStream"
      @close="showMirrorModal = false"
      @created="onStreamCreated"
    />

    <!-- Duplicate modal -->
    <StreamCreateModal
      v-if="showDuplicateModal && streamsStore.selectedStream"
      :duplicate-stream="streamsStore.selectedStream"
      @close="showDuplicateModal = false"
      @created="onStreamCreated"
    />

    <!-- Purge modal -->
    <Teleport to="body">
      <div v-if="showPurgeModal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/50" @click="showPurgeModal = false"></div>
        <div class="relative w-full max-w-md bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4 p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">Purge Stream</h3>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">Remove all messages from <strong class="text-gray-700 dark:text-gray-300">{{ selectedStreamName }}</strong>. Leave subject empty to purge all messages.</p>
          <div class="mb-3">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subject filter (optional)</label>
            <input
              v-model="purgeSubject"
              type="text"
              placeholder="e.g. orders.>"
              class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
            />
          </div>
          <p v-if="purgeError" class="mb-2 text-xs text-red-500">{{ purgeError }}</p>
          <div class="flex justify-end gap-3">
            <button class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="showPurgeModal = false">Cancel</button>
            <button class="px-4 py-2 text-sm font-medium text-white bg-amber-600 rounded-md hover:bg-amber-700" @click="handlePurge">Purge Stream</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
