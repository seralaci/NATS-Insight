<script setup lang="ts">
import { useUiStore } from '../stores/ui'
import ConnectionDropdown from '../features/connections/ConnectionDropdown.vue'

const uiStore = useUiStore()
</script>

<template>
  <header class="h-12 border-b border-gray-200 dark:border-gray-800 flex items-center px-3 gap-2 bg-white dark:bg-gray-950 shrink-0">
    <!-- Logo -->
    <button
      class="flex items-center gap-2 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800"
      @click="uiStore.toggleSidebar()"
      title="Toggle sidebar"
    >
      <svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
      </svg>
      <span class="text-sm font-semibold text-gray-900 dark:text-gray-100">NATS Insight</span>
    </button>

    <!-- Actions -->
    <div class="flex items-center gap-1 ml-2">
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400"
        title="Create new connection"
        @click="$emit('create-connection')"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
        </svg>
      </button>
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400"
        title="Manage connections"
        @click="$emit('manage-connections')"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75"/>
        </svg>
      </button>
    </div>

    <!-- Connection dropdown -->
    <div class="flex-1 max-w-xl mx-4">
      <ConnectionDropdown @create-connection="$emit('create-connection')" />
    </div>

    <!-- Right side: theme toggle -->
    <div class="flex items-center gap-1 ml-auto">
      <button
        class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-500 dark:text-gray-400"
        title="Toggle theme"
        @click="uiStore.setTheme(uiStore.theme === 'dark' ? 'light' : 'dark')"
      >
        <!-- Sun icon (dark mode) -->
        <svg v-if="uiStore.theme === 'dark'" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z"/>
        </svg>
        <!-- Moon icon (light mode) -->
        <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z"/>
        </svg>
      </button>
    </div>
  </header>
</template>
