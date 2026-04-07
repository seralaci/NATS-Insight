<script setup lang="ts">
import { useConnectionsStore } from '../../stores/connections'
import type { Connection } from '../../lib/api'

const emit = defineEmits<{
  close: []
  edit: [Connection]
  create: []
}>()

const connStore = useConnectionsStore()

const authLabel: Record<string, string> = {
  none: 'No Auth',
  username_password: 'Username/PW',
  token: 'Token',
  nkey: 'NKey',
  credentials: 'Creds',
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="$emit('close')"></div>

      <div class="relative w-full max-w-lg max-h-[70vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Manage Connections</h2>
          <button
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            @click="$emit('close')"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Connection list -->
        <div class="flex-1 overflow-y-auto">
          <div v-if="connStore.connections.length === 0" class="p-8 text-center text-sm text-gray-500 dark:text-gray-400">
            No connections yet.
          </div>
          <button
            v-for="conn in connStore.connections"
            :key="conn.id"
            class="w-full flex items-center gap-3 px-6 py-3 text-sm hover:bg-gray-50 dark:hover:bg-gray-900 border-b border-gray-100 dark:border-gray-800 text-left"
            @click="$emit('edit', conn)"
          >
            <div class="flex-1">
              <div class="font-medium text-gray-900 dark:text-gray-100">{{ conn.name }}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ conn.url }}</div>
            </div>
            <span class="text-[10px] px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 shrink-0">
              {{ authLabel[conn.authMethod] || conn.authMethod }}
            </span>
            <svg class="w-4 h-4 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
            </svg>
          </button>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0">
          <button
            class="w-full py-2.5 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700"
            @click="$emit('create')"
          >
            + Create New Connection
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
