<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useStreamsStore } from '../../stores/streams'
import type { ConsumerInfo } from '../../lib/api'

const props = defineProps<{
  streamName: string
}>()

const emit = defineEmits<{
  selectConsumer: [consumer: ConsumerInfo]
}>()

const streamsStore = useStreamsStore()
const searchQuery = ref('')

const filteredConsumers = computed(() => {
  const q = searchQuery.value.toLowerCase().trim()
  if (!q) return streamsStore.consumers
  return streamsStore.consumers.filter(c => c.name.toLowerCase().includes(q))
})

function filterSubjectsDisplay(config: any): string {
  if (!config) return '---'
  if (config.filter_subject) return config.filter_subject
  if (config.filter_subjects?.length) return config.filter_subjects.join(', ')
  return '>'
}

async function refresh() {
  await streamsStore.fetchConsumers(props.streamName)
}

onMounted(() => refresh())
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Toolbar -->
    <div class="flex items-center gap-2 px-4 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0">
      <!-- Refresh -->
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 disabled:opacity-40"
        :disabled="streamsStore.consumersLoading"
        title="Refresh"
        @click="refresh"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
        </svg>
      </button>

      <!-- Search filter -->
      <div class="flex-1 relative">
        <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z"/>
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="`Filter consumers (${streamsStore.consumers.length})`"
          class="w-full pl-8 pr-3 py-1 text-xs border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-600 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
        />
      </div>
    </div>

    <!-- Table -->
    <div class="flex-1 overflow-auto">
      <div v-if="streamsStore.consumersLoading" class="flex items-center justify-center py-20">
        <span class="text-sm text-gray-400 dark:text-gray-600">Loading consumers...</span>
      </div>

      <div v-else-if="streamsStore.consumers.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
        <svg class="w-10 h-10 text-gray-300 dark:text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
          <path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 0 0 3.741-.479 3 3 0 0 0-4.682-2.72m.94 3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0 1 12 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0 1 6 18.719m12 0a5.971 5.971 0 0 0-.941-3.197m0 0A5.995 5.995 0 0 0 12 12.75a5.995 5.995 0 0 0-5.058 2.772m0 0a3 3 0 0 0-4.681 2.72 8.986 8.986 0 0 0 3.74.477m.94-3.197a5.971 5.971 0 0 0-.94 3.197M15 6.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm6 3a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Zm-13.5 0a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Z"/>
        </svg>
        <span class="text-sm text-gray-500 dark:text-gray-400">No consumers found</span>
      </div>

      <div v-else-if="filteredConsumers.length === 0" class="flex items-center justify-center py-20">
        <span class="text-sm text-gray-400 dark:text-gray-600">No consumers matching "{{ searchQuery }}"</span>
      </div>

      <table v-else class="w-full text-sm">
        <thead class="sticky top-0 bg-white dark:bg-gray-950 border-b border-gray-200 dark:border-gray-800">
          <tr>
            <th class="w-8 px-3 py-2">
              <input type="checkbox" class="rounded border-gray-300 dark:border-gray-700 text-emerald-600" disabled />
            </th>
            <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Name</th>
            <th class="text-left px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Filter Subjects</th>
            <th class="text-right px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Unprocessed Msgs</th>
            <th class="text-right px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Outstanding Acks</th>
            <th class="text-right px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Ack Floor</th>
            <th class="text-right px-4 py-2 text-xs font-medium text-gray-500 dark:text-gray-400">Waiting Clients</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="c in filteredConsumers"
            :key="c.name"
            class="border-b border-gray-100 dark:border-gray-800/50 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/40"
            @click="emit('selectConsumer', c)"
          >
            <td class="px-3 py-2.5" @click.stop>
              <input type="checkbox" class="rounded border-gray-300 dark:border-gray-700 text-emerald-600" />
            </td>
            <td class="px-4 py-2.5">
              <div class="flex items-center gap-2">
                <!-- Paused indicator icon -->
                <svg
                  v-if="c.paused"
                  class="w-3.5 h-3.5 text-yellow-500 shrink-0"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5"/>
                </svg>
                <svg
                  v-else
                  class="w-3.5 h-3.5 text-green-500 shrink-0"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z"/>
                </svg>
                <span class="font-mono text-xs text-gray-800 dark:text-gray-200">{{ c.name }}</span>
              </div>
            </td>
            <td class="px-4 py-2.5 font-mono text-xs text-emerald-600 dark:text-emerald-400">{{ filterSubjectsDisplay(c.config) }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ c.num_pending }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ c.num_ack_pending }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ c.ack_floor?.stream_seq ?? 0 }}</td>
            <td class="px-4 py-2.5 font-mono text-xs text-right text-gray-700 dark:text-gray-300">{{ c.num_waiting }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
