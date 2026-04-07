<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { useObjectsStore } from '../../stores/objects'

ChartJS.register(ArcElement, Tooltip, Legend)

const objectsStore = useObjectsStore()

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

const topBySize = computed(() => {
  return [...objectsStore.stores]
    .sort((a, b) => b.size - a.size)
    .slice(0, 5)
})

const maxSize = computed(() => topBySize.value[0]?.size ?? 1)

const fileBytes = computed(() =>
  objectsStore.stores
    .filter(s => s.storage === 'file')
    .reduce((sum, s) => sum + s.size, 0)
)

const memoryBytes = computed(() =>
  objectsStore.stores
    .filter(s => s.storage === 'memory')
    .reduce((sum, s) => sum + s.size, 0)
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

    <!-- Object Buckets by Size -->
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Object Buckets by Size</h3>
      <div v-if="topBySize.length > 0">
        <div class="flex justify-between mb-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Name</span>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Size</span>
        </div>
        <div class="space-y-3">
          <div v-for="store in topBySize" :key="store.bucket">
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100 truncate mr-4 max-w-[60%]">{{ store.bucket }}</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400 shrink-0">{{ formatBytes(store.size) }}</span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-500"
                :style="{ width: maxSize > 0 ? (store.size / maxSize * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
      <p v-else class="text-xs text-gray-400 dark:text-gray-600">No object buckets</p>
    </div>

    <!-- Object Buckets by Storage -->
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3">Object Buckets by Storage</h3>
      <div v-if="hasStorageData" class="relative h-48">
        <Doughnut :data="doughnutData" :options="doughnutOptions" />
      </div>
      <div v-else class="flex flex-col gap-3">
        <div class="flex justify-between mb-2">
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Type</span>
          <span class="text-xs font-medium text-gray-500 dark:text-gray-400">Buckets</span>
        </div>
        <div class="space-y-3">
          <div>
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100">File</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400">
                {{ objectsStore.stores.filter(s => s.storage === 'file').length.toLocaleString() }}
              </span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-500"
                :style="{ width: objectsStore.stores.length > 0 ? (objectsStore.stores.filter(s => s.storage === 'file').length / objectsStore.stores.length * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
          <div>
            <div class="flex justify-between items-center mb-1">
              <span class="text-xs font-mono text-gray-900 dark:text-gray-100">Memory</span>
              <span class="text-xs font-mono text-gray-500 dark:text-gray-400">
                {{ objectsStore.stores.filter(s => s.storage === 'memory').length.toLocaleString() }}
              </span>
            </div>
            <div class="h-1.5 rounded-full bg-gray-100 dark:bg-gray-800">
              <div
                class="h-1.5 rounded-full bg-emerald-300"
                :style="{ width: objectsStore.stores.length > 0 ? (objectsStore.stores.filter(s => s.storage === 'memory').length / objectsStore.stores.length * 100) + '%' : '0%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>
