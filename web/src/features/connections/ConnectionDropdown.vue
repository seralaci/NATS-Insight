<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useConnectionsStore } from '../../stores/connections'

const emit = defineEmits<{
  'create-connection': []
}>()

const connStore = useConnectionsStore()
const isOpen = ref(false)
const search = ref('')
const dropdownRef = ref<HTMLElement>()
const showInfo = ref(false)

const filteredConnections = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return connStore.connections
  return connStore.connections.filter(c =>
    c.name.toLowerCase().includes(q) || c.url.toLowerCase().includes(q)
  )
})

const authLabel: Record<string, string> = {
  none: 'No Auth',
  username_password: 'Username/PW',
  token: 'Token',
  nkey: 'NKey',
  credentials: 'Creds',
}

function toggle() {
  if (showInfo.value) {
    showInfo.value = false
    return
  }
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    search.value = ''
    showInfo.value = false
  }
}

function toggleInfo() {
  showInfo.value = !showInfo.value
  isOpen.value = false
}

async function selectConnection(id: string) {
  isOpen.value = false
  await connStore.connect(id)
}

async function handleDisconnect() {
  await connStore.disconnect()
  showInfo.value = false
}

function handleClickOutside(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
    showInfo.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<template>
  <div ref="dropdownRef" class="relative">
    <!-- Trigger button -->
    <div class="flex items-center gap-1">
      <button
        class="flex-1 flex items-center gap-2 px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-300 hover:border-gray-400 dark:hover:border-gray-600 "
        @click="toggle"
      >
        <span
          v-if="connStore.activeConnection"
          class="w-2 h-2 rounded-full shrink-0"
          :class="connStore.status.connected ? 'bg-green-500' : 'bg-yellow-500'"
        ></span>

        <span class="flex-1 text-left truncate">
          {{ connStore.activeConnection?.name || 'Select a connection...' }}
        </span>

        <svg class="w-4 h-4 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5"/>
        </svg>
      </button>

      <!-- Info button (only when connected) -->
      <button
        v-if="connStore.status.connected"
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 "
        title="Connection info"
        @click.stop="toggleInfo"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"/>
        </svg>
      </button>
    </div>

    <!-- Connection info popup -->
    <div
      v-if="showInfo && connStore.status.connected"
      class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-md shadow-lg z-50 overflow-hidden"
    >
      <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <span class="text-sm font-medium text-gray-900 dark:text-gray-100">Connection Info</span>
        <button
          class="text-xs px-2 py-1 rounded text-red-500 hover:bg-red-50 dark:hover:bg-red-950 "
          @click="handleDisconnect"
        >
          Disconnect
        </button>
      </div>
      <table class="w-full text-sm">
        <tbody>
          <tr class="border-b border-gray-100 dark:border-gray-800">
            <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">Status</td>
            <td class="px-4 py-1.5 text-green-600 dark:text-green-400 font-medium">Connected</td>
          </tr>
          <tr class="border-b border-gray-100 dark:border-gray-800">
            <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">Auth Method</td>
            <td class="px-4 py-1.5 text-gray-900 dark:text-gray-100">{{ authLabel[connStore.activeConnection?.authMethod ?? 'none'] }}</td>
          </tr>
          <tr v-if="connStore.status.serverName" class="border-b border-gray-100 dark:border-gray-800">
            <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">Server Name</td>
            <td class="px-4 py-1.5 text-gray-900 dark:text-gray-100 font-mono text-xs truncate">{{ connStore.status.serverName }}</td>
          </tr>
          <tr v-if="connStore.status.clusterName" class="border-b border-gray-100 dark:border-gray-800">
            <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">Cluster</td>
            <td class="px-4 py-1.5 text-gray-900 dark:text-gray-100 font-mono text-xs truncate">{{ connStore.status.clusterName }}</td>
          </tr>
          <tr v-if="connStore.status.serverId" class="border-b border-gray-100 dark:border-gray-800">
            <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">Server ID</td>
            <td class="px-4 py-1.5 text-gray-900 dark:text-gray-100 font-mono text-xs truncate">{{ connStore.status.serverId }}</td>
          </tr>
          <tr v-if="connStore.status.version" class="border-b border-gray-100 dark:border-gray-800">
            <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">Version</td>
            <td class="px-4 py-1.5 text-gray-900 dark:text-gray-100">{{ connStore.status.version }}</td>
          </tr>
          <tr>
            <td class="px-4 py-1.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">RTT</td>
            <td class="px-4 py-1.5 text-gray-900 dark:text-gray-100">{{ connStore.status.rtt || '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Connection list dropdown -->
    <div
      v-if="isOpen"
      class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-md shadow-lg z-50 overflow-hidden"
    >
      <!-- Search -->
      <div class="p-2 border-b border-gray-200 dark:border-gray-700">
        <input
          v-model="search"
          type="text"
          placeholder="Search connections..."
          class="w-full px-2 py-1 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded text-gray-900 dark:text-gray-100 placeholder-gray-400 focus:outline-none focus:ring-1 focus:ring-emerald-500"
          @click.stop
        />
      </div>

      <!-- Connection list -->
      <div class="max-h-60 overflow-y-auto">
        <button
          v-if="filteredConnections.length === 0"
          class="w-full p-3 text-sm text-emerald-600 dark:text-emerald-400 hover:bg-gray-100 dark:hover:bg-gray-800 text-center cursor-pointer "
          @click="isOpen = false; $emit('create-connection')"
        >
          No connections found. <span class="underline">Create a new connection</span>
        </button>
        <button
          v-for="conn in filteredConnections"
          :key="conn.id"
          class="w-full flex items-center gap-3 px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-800 "
          :class="connStore.status.connectionId === conn.id ? 'bg-gray-50 dark:bg-gray-800/50' : ''"
          @click="selectConnection(conn.id)"
        >
          <div class="flex-1 text-left">
            <div class="font-medium text-gray-900 dark:text-gray-100">{{ conn.name }}</div>
            <div class="text-xs text-gray-500 dark:text-gray-400">{{ conn.url }}</div>
          </div>
          <span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400">
            {{ authLabel[conn.authMethod] || conn.authMethod }}
          </span>
        </button>
      </div>
    </div>
  </div>
</template>
