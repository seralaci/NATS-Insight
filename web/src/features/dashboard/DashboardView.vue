<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useConnectionsStore } from '../../stores/connections'
import { useServerStore } from '../../stores/server'
import { useStreamsStore } from '../../stores/streams'
import { useKvStore } from '../../stores/kv'
import { useObjectsStore } from '../../stores/objects'
import { useMetrics } from '../../composables/useMetrics'
import WelcomeView from '../welcome/WelcomeView.vue'
import MetricCard from './MetricCard.vue'
import MetricChart from './MetricChart.vue'
import StreamCharts from './StreamCharts.vue'
import KvCharts from './KvCharts.vue'
import ObjectCharts from './ObjectCharts.vue'
import ConnectionsTable from './ConnectionsTable.vue'
import RefreshControl from './RefreshControl.vue'

const connStore = useConnectionsStore()
const serverStore = useServerStore()
const streamsStore = useStreamsStore()
const kvStore = useKvStore()
const objectsStore = useObjectsStore()
const metrics = useMetrics(60)

const refreshInterval = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

function setRefreshInterval(seconds: number) {
  refreshInterval.value = seconds
  if (timer) clearInterval(timer)
  timer = null
  if (seconds > 0) {
    timer = setInterval(() => serverStore.fetchData(), seconds * 1000)
  }
}

const isConnected = computed(() => connStore.status.connected)

watch(isConnected, (connected) => {
  if (connected) {
    serverStore.fetchData()
    metrics.start(5)
  } else {
    serverStore.clear()
    metrics.stop()
  }
})

onMounted(() => {
  if (connStore.status.connected) {
    serverStore.fetchData()
    metrics.start(5)
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>

<template>
  <WelcomeView v-if="!connStore.status.connected" />

  <div v-else class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">Dashboard</h1>
        <span
          v-if="metrics.connected.value"
          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-950/60 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-green-500 shrink-0"></span>
          Live
        </span>
      </div>
      <RefreshControl :interval="refreshInterval" @change="setRefreshInterval" @refresh="serverStore.fetchData()" />
    </div>

    <div v-if="serverStore.error" class="p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800">
      {{ serverStore.error }}
    </div>

    <!-- Streams -->
    <section>
      <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Streams</h2>
      <div class="grid grid-cols-4 gap-4">
        <MetricCard label="Streams" :value="serverStore.jsz?.streams ?? 0" />
        <MetricCard label="Messages" :value="serverStore.jsz?.messages ?? 0" />
        <MetricCard label="Total Size" :value="formatBytes(serverStore.jsz?.bytes ?? 0)" />
        <MetricCard label="Consumers" :value="serverStore.jsz?.consumers ?? 0" />
      </div>
      <StreamCharts v-if="streamsStore.streams.length > 0" class="mt-4" />
    </section>

    <!-- KV -->
    <section>
      <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">KV</h2>
      <div class="grid grid-cols-3 gap-4">
        <MetricCard label="KV Buckets" :value="kvStore.buckets.length" />
        <MetricCard label="Size" :value="formatBytes(kvStore.buckets.reduce((sum, b) => sum + b.bytes, 0))" />
        <MetricCard label="Values" :value="kvStore.buckets.reduce((sum, b) => sum + b.values, 0).toLocaleString()" />
      </div>
      <KvCharts v-if="kvStore.buckets.length > 0" class="mt-4" />
    </section>

    <!-- Object Buckets -->
    <section>
      <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Object Buckets</h2>
      <div class="grid grid-cols-3 gap-4">
        <MetricCard label="Object Buckets" :value="objectsStore.stores.length" />
        <MetricCard label="Total Size" :value="formatBytes(objectsStore.stores.reduce((sum, s) => sum + s.size, 0))" />
        <MetricCard label="Sealed Object Buckets" :value="objectsStore.stores.filter(s => s.sealed).length" />
      </div>
      <ObjectCharts v-if="objectsStore.stores.length > 0" class="mt-4" />
    </section>

    <!-- Server -->
    <section v-if="serverStore.varz">
      <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Server</h2>
      <div class="grid grid-cols-4 gap-4">
        <MetricCard label="Version" :value="serverStore.varz.version" />
        <MetricCard label="Uptime" :value="serverStore.varz.uptime" />
        <MetricCard label="Connections" :value="serverStore.varz.connections" />
        <MetricCard label="Subscriptions" :value="serverStore.varz.subscriptions" />
      </div>
    </section>

    <!-- Real-time Charts -->
    <section>
      <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">Real-time Metrics</h2>
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <MetricChart
          title="Messages / sec"
          :points="metrics.points.value"
          :data-keys="['msgsIn', 'msgsOut']"
          :labels="['In', 'Out']"
          :colors="['#6366f1', '#22c55e']"
          unit="msg/s"
        />
        <MetricChart
          title="Data / sec"
          :points="metrics.points.value"
          :data-keys="['bytesIn', 'bytesOut']"
          :labels="['In', 'Out']"
          :colors="['#a855f7', '#f97316']"
          unit="B/s"
        />
        <MetricChart
          title="Active Connections"
          :points="metrics.points.value"
          :data-keys="['connections']"
          :labels="['Connections']"
          :colors="['#6366f1']"
        />
      </div>
    </section>

    <!-- Connections Table -->
    <section>
      <ConnectionsTable />
    </section>
  </div>
</template>
