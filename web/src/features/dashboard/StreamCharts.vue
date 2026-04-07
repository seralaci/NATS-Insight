<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { useStreamsStore } from '../../stores/streams'

ChartJS.register(ArcElement, Tooltip, Legend)

const streamsStore = useStreamsStore()

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const topByMessages = computed(() => {
  return [...streamsStore.streams]
    .sort((a, b) => b.state.messages - a.state.messages)
    .slice(0, 5)
})

const topByBytes = computed(() => {
  return [...streamsStore.streams]
    .sort((a, b) => b.state.bytes - a.state.bytes)
    .slice(0, 5)
})

const topByConsumers = computed(() => {
  return [...streamsStore.streams]
    .sort((a, b) => b.state.consumer_count - a.state.consumer_count)
    .slice(0, 5)
})

const maxMessages = computed(() => topByMessages.value[0]?.state.messages ?? 1)
const maxBytes = computed(() => topByBytes.value[0]?.state.bytes ?? 1)
const maxConsumers = computed(() => topByConsumers.value[0]?.state.consumer_count ?? 1)

const fileBytes = computed(() =>
  streamsStore.streams
    .filter(s => s.config.storage === 'file')
    .reduce((sum, s) => sum + s.state.bytes, 0)
)

const memoryBytes = computed(() =>
  streamsStore.streams
    .filter(s => s.config.storage === 'memory')
    .reduce((sum, s) => sum + s.state.bytes, 0)
)

const hasStorageData = computed(() => fileBytes.value > 0 || memoryBytes.value > 0)

const isDark = computed(() => document.documentElement.classList.contains('dark'))

const doughnutData = computed(() => ({
  labels: ['File', 'Memory'],
  datasets: [
    {
      data: [fileBytes.value, memoryBytes.value],
      backgroundColor: ['#10b981', '#6ee7b7'],
      borderWidth: 0,
    },
  ],
}))

const doughnutOptions = computed(() => ({
  cutout: '65%',
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        color: isDark.value ? '#9ca3af' : '#6b7280',
        font: { size: 12 },
        padding: 16,
        usePointStyle: true,
        pointStyleWidth: 8,
      },
    },
    tooltip: {
      callbacks: {
        label: (ctx: any) => {
          const label = ctx.label ?? ''
          const value = formatBytes(ctx.parsed)
          return ` ${label}: ${value}`
        },
      },
    },
  },
}))
</script>

<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

    <!-- Streams by Message Amount -->
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Streams by Message Amount</h3>
      <div v-if="topByMessages.length > 0">
        <div class="flex justify-between mb-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Name</span>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Messages</span>
        </div>
        <div class="space-y-3">
          <div v-for="stream in topByMessages" :key="stream.config.name">
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100 truncate mr-4 max-w-[60%]">{{ stream.config.name }}</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400 shrink-0">{{ stream.state.messages.toLocaleString() }}</span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-500"
                :style="{ width: maxMessages > 0 ? (stream.state.messages / maxMessages * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
      <p v-else class="text-xs text-gray-400 dark:text-gray-600">No streams</p>
    </div>

    <!-- Streams by Message Size -->
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Streams by Message Size</h3>
      <div v-if="topByBytes.length > 0">
        <div class="flex justify-between mb-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Name</span>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Size</span>
        </div>
        <div class="space-y-3">
          <div v-for="stream in topByBytes" :key="stream.config.name">
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100 truncate mr-4 max-w-[60%]">{{ stream.config.name }}</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400 shrink-0">{{ formatBytes(stream.state.bytes) }}</span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-500"
                :style="{ width: maxBytes > 0 ? (stream.state.bytes / maxBytes * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
      <p v-else class="text-xs text-gray-400 dark:text-gray-600">No streams</p>
    </div>

    <!-- Streams by Consumers -->
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Streams by Consumers</h3>
      <div v-if="topByConsumers.length > 0">
        <div class="flex justify-between mb-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Name</span>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Consumers</span>
        </div>
        <div class="space-y-3">
          <div v-for="stream in topByConsumers" :key="stream.config.name">
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100 truncate mr-4 max-w-[60%]">{{ stream.config.name }}</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400 shrink-0">{{ stream.state.consumer_count.toLocaleString() }}</span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-500"
                :style="{ width: maxConsumers > 0 ? (stream.state.consumer_count / maxConsumers * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
      <p v-else class="text-xs text-gray-400 dark:text-gray-600">No streams</p>
    </div>

    <!-- Streams by Storage -->
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Streams by Storage</h3>
      <div v-if="hasStorageData" class="relative h-48">
        <Doughnut :data="doughnutData" :options="doughnutOptions" />
      </div>
      <div v-else class="flex flex-col gap-3">
        <div class="flex justify-between mb-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Type</span>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Streams</span>
        </div>
        <div class="space-y-3">
          <div>
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100">File</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400">
                {{ streamsStore.streams.filter(s => s.config.storage === 'file').length.toLocaleString() }}
              </span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-500"
                :style="{ width: streamsStore.streams.length > 0 ? (streamsStore.streams.filter(s => s.config.storage === 'file').length / streamsStore.streams.length * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
          <div>
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100">Memory</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400">
                {{ streamsStore.streams.filter(s => s.config.storage === 'memory').length.toLocaleString() }}
              </span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-300"
                :style="{ width: streamsStore.streams.length > 0 ? (streamsStore.streams.filter(s => s.config.storage === 'memory').length / streamsStore.streams.length * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>
