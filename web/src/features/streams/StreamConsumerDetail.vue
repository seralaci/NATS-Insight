<script setup lang="ts">
import { ref } from 'vue'
import { useStreamsStore } from '../../stores/streams'
import type { ConsumerInfo } from '../../lib/api'

const props = defineProps<{
  consumer: ConsumerInfo
  streamName: string
}>()

const emit = defineEmits<{
  back: []
  refresh: []
}>()

const streamsStore = useStreamsStore()
const showPauseModal = ref(false)
const pauseUntilInput = ref('')
const pauseError = ref('')
const confirmDelete = ref(false)

function val(v: any, fallback = '---'): string {
  if (v === undefined || v === null || v === '') return fallback
  return String(v)
}

function boolVal(v: any): string {
  if (v === undefined || v === null) return '---'
  return v ? 'true' : 'false'
}

function formatDate(iso: string) {
  if (!iso) return '---'
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function formatNs(ns: number): string {
  if (!ns || ns <= 0) return '---'
  const secs = ns / 1e9
  if (secs < 60) return `${secs.toFixed(1)}s`
  if (secs < 3600) return `${Math.round(secs / 60)}m`
  return `${Math.round(secs / 3600)}h`
}

function filterSubjectsDisplay(config: any): string {
  if (!config) return '---'
  if (config.filterSubject) return config.filterSubject
  if (config.filterSubjects?.length) return config.filterSubjects.join(', ')
  return '>'
}

function consumerType(config: any): string {
  if (!config) return '---'
  if (config.durable || config.durableName) return 'Durable'
  return 'Ephemeral'
}

function consumerMode(config: any): string {
  if (!config) return '---'
  if (config.pushBound !== undefined) return 'Push'
  return 'Pull'
}

async function openPause() {
  pauseUntilInput.value = ''
  pauseError.value = ''
  showPauseModal.value = true
}

async function submitPause() {
  if (!pauseUntilInput.value) {
    pauseError.value = 'Please enter a date/time'
    return
  }
  pauseError.value = ''
  try {
    await streamsStore.pauseConsumer(props.streamName, props.consumer.name, pauseUntilInput.value)
    showPauseModal.value = false
    emit('refresh')
  } catch (e: any) {
    pauseError.value = e.message
  }
}

async function doResume() {
  await streamsStore.resumeConsumer(props.streamName, props.consumer.name)
  emit('refresh')
}

async function doDelete() {
  if (!confirmDelete.value) {
    confirmDelete.value = true
    return
  }
  await streamsStore.deleteConsumer(props.streamName, props.consumer.name)
  emit('back')
}
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <!-- Header -->
    <div class="flex items-center gap-3 px-4 py-2.5 border-b border-gray-200 dark:border-gray-800 shrink-0">
      <button
        class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
        @click="emit('back')"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5"/>
        </svg>
        Consumers
      </button>
      <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
      </svg>
      <span class="text-xs font-mono font-semibold text-gray-700 dark:text-gray-300">{{ streamName }}</span>

      <div class="ml-auto flex items-center gap-2">
        <!-- Pause / Resume -->
        <button
          v-if="!consumer.paused"
          class="px-3 py-1 text-xs border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
          @click="openPause"
        >Pause</button>
        <button
          v-else
          class="px-3 py-1 text-xs bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
          @click="doResume"
        >Resume</button>
        <!-- Delete -->
        <button
          class="px-3 py-1 text-xs rounded-md"
          :class="confirmDelete
            ? 'bg-red-600 text-white hover:bg-red-700'
            : 'text-red-500 border border-red-300 dark:border-red-800 hover:bg-red-50 dark:hover:bg-red-950'"
          @click="doDelete"
        >{{ confirmDelete ? 'Confirm Delete' : 'Delete' }}</button>
      </div>
    </div>

    <!-- Consumer name bar -->
    <div class="flex items-center gap-3 px-4 py-2 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/40 shrink-0">
      <svg class="w-4 h-4 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 0 0 3.741-.479 3 3 0 0 0-4.682-2.72m.94 3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0 1 12 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0 1 6 18.719m12 0a5.971 5.971 0 0 0-.941-3.197m0 0A5.995 5.995 0 0 0 12 12.75a5.995 5.995 0 0 0-5.058 2.772m0 0a3 3 0 0 0-4.681 2.72 8.986 8.986 0 0 0 3.74.477m.94-3.197a5.971 5.971 0 0 0-.94 3.197M15 6.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm6 3a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Zm-13.5 0a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Z"/>
      </svg>
      <span class="text-sm font-mono font-semibold text-gray-800 dark:text-gray-200">{{ consumer.name }}</span>
      <span v-if="consumer.paused" class="ml-1 px-1.5 py-0.5 text-[10px] bg-yellow-100 dark:bg-yellow-900/40 text-yellow-700 dark:text-yellow-400 rounded">paused</span>
      <span class="text-xs text-gray-400 dark:text-gray-600 ml-2">Created {{ formatDate(consumer.created) }}</span>
    </div>

    <!-- Detail content -->
    <div class="flex-1 overflow-auto p-4">
      <!-- Top row: 3 columns -->
      <div class="grid grid-cols-3 gap-4 mb-4">

        <!-- General Configuration -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">General Configuration</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Name</span>
              <span class="font-mono text-gray-800 dark:text-gray-200 text-right">{{ consumer.name }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Stream</span>
              <span class="font-mono text-emerald-600 dark:text-emerald-400">{{ streamName }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Type</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumerType(consumer.config) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Mode</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumerMode(consumer.config) }}</span>
            </div>
            <div v-if="consumer.config?.description" class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Description</span>
              <span class="text-gray-700 dark:text-gray-300 text-right max-w-[10rem] truncate">{{ consumer.config.description }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Filter Subject</span>
              <span class="font-mono text-emerald-600 dark:text-emerald-400 text-right max-w-[10rem] truncate">{{ val(consumer.config?.filterSubject) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Filter Subjects</span>
              <span class="font-mono text-emerald-600 dark:text-emerald-400 text-right max-w-[10rem] truncate">{{ filterSubjectsDisplay(consumer.config) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Pause Status</span>
              <span class="font-mono" :class="consumer.paused ? 'text-yellow-600 dark:text-yellow-400' : 'text-green-600 dark:text-green-400'">
                {{ consumer.paused ? 'Paused' : 'Not Paused' }}
              </span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Ack Policy</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.ackPolicy || consumer.config?.ack_policy) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Ack Wait</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatNs(consumer.config?.ackWait || consumer.config?.ack_wait) }}</span>
            </div>
          </div>
        </div>

        <!-- Delivery Settings -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Delivery Settings</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Deliver Policy</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.deliverPolicy || consumer.config?.deliver_policy) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Replay Policy</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.replayPolicy || consumer.config?.replay_policy) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Max Deliveries</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">
                {{ (consumer.config?.maxDeliver === -1 || consumer.config?.max_deliver === -1) ? 'Unlimited' : val(consumer.config?.maxDeliver || consumer.config?.max_deliver) }}
              </span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Max Ack Pending</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.maxAckPending || consumer.config?.max_ack_pending) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Backoff</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.config?.backoff?.length ? consumer.config.backoff.length + ' steps' : '---' }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Sample Frequency</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.sampleFrequency || consumer.config?.sample_freq) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Memory Storage</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ boolVal(consumer.config?.memStorage || consumer.config?.mem_storage) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Inactive Threshold</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatNs(consumer.config?.inactiveThreshold || consumer.config?.inactive_threshold) }}</span>
            </div>
          </div>
        </div>

        <!-- Pull Configuration -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Pull Configuration</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Max Waiting Pull Requests</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.maxWaiting || consumer.config?.max_waiting) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Max Request Expires</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ formatNs(consumer.config?.maxExpires || consumer.config?.max_expires) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Max Request Batch</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.maxBatch || consumer.config?.max_batch) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Max Request Max Bytes</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.maxBytes || consumer.config?.max_bytes) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom row: 3 columns -->
      <div class="grid grid-cols-3 gap-4">

        <!-- Delivery State -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Delivery State</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Last Delivered (Stream Seq)</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.delivered?.stream_seq ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Last Delivered (Consumer Seq)</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.delivered?.consumer_seq ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Ack Floor (Stream Seq)</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.ack_floor?.stream_seq ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Ack Floor (Consumer Seq)</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.ack_floor?.consumer_seq ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Last Active</span>
              <span class="font-mono text-gray-700 dark:text-gray-300 text-right">{{ val(consumer.config?.lastActive) }}</span>
            </div>
          </div>
        </div>

        <!-- Consumer Status -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Consumer Status</h3>
          <div class="space-y-2">
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Redelivered Messages</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.num_redelivered ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Unprocessed Messages</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.num_pending ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Outstanding Acks</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.num_ack_pending ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Waiting Clients</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ consumer.num_waiting ?? 0 }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Cluster</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.cluster) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Leader</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.leader) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Desired Replicas</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.numReplicas || consumer.config?.num_replicas) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="text-gray-500 dark:text-gray-400">Current Replicas</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">---</span>
            </div>
          </div>
        </div>

        <!-- Metadata -->
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Metadata</h3>
          <!-- User Metadata -->
          <h4 class="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-600 mb-2">User Metadata</h4>
          <div
            v-if="consumer.config?.metadata && Object.keys(consumer.config.metadata).length > 0"
            class="space-y-1.5 mb-3"
          >
            <div
              v-for="(mval, mkey) in consumer.config.metadata"
              :key="mkey"
              class="flex justify-between text-xs"
            >
              <span class="font-mono text-gray-500 dark:text-gray-400">{{ mkey }}</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ mval }}</span>
            </div>
          </div>
          <div v-else class="text-xs text-gray-400 dark:text-gray-600 italic mb-3">No user-defined metadata</div>

          <!-- NATS Metadata -->
          <h4 class="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-600 mb-2">NATS Metadata</h4>
          <div class="space-y-1.5">
            <div class="flex justify-between text-xs">
              <span class="font-mono text-gray-500 dark:text-gray-400">_nats.level</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.metadata?.['_nats.level']) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="font-mono text-gray-500 dark:text-gray-400">_nats.msg.level</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.metadata?.['_nats.msg.level']) }}</span>
            </div>
            <div class="flex justify-between text-xs">
              <span class="font-mono text-gray-500 dark:text-gray-400">_nats.snr</span>
              <span class="font-mono text-gray-800 dark:text-gray-200">{{ val(consumer.config?.metadata?.['_nats.snr']) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Pause modal -->
  <Teleport to="body">
    <div v-if="showPauseModal" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="showPauseModal = false"></div>
      <div class="relative w-full max-w-md bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4 p-6">
        <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">Pause Consumer</h3>
        <p class="text-sm text-gray-500 dark:text-gray-400 mb-3">Pause consumer until a specific date/time:</p>
        <input
          v-model="pauseUntilInput"
          type="datetime-local"
          class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
        />
        <p v-if="pauseError" class="mt-2 text-xs text-red-500">{{ pauseError }}</p>
        <div class="flex justify-end gap-3 mt-4">
          <button class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="showPauseModal = false">Cancel</button>
          <button class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700" @click="submitPause">Pause</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
