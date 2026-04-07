<script setup lang="ts">
import { ref, computed } from 'vue'
import type { WsStatus } from '../../lib/ws'

const props = defineProps<{
  isRunning: boolean
  isPaused: boolean
  messageCount: number
  maxBuffer: number
  status: WsStatus
}>()

const emit = defineEmits<{
  start: [subject: string]
  stop: []
  pause: []
  resume: []
  clear: []
  publish: []
}>()

const subjectInput = ref('')

function handleStart() {
  const s = subjectInput.value.trim()
  if (!s) return
  emit('start', s)
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !props.isRunning) {
    handleStart()
  }
}

const bufferPercent = computed(() =>
  props.maxBuffer > 0 ? (props.messageCount / props.maxBuffer) * 100 : 0
)
</script>

<template>
  <div class="flex items-center gap-2 px-4 py-3 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-950 flex-wrap">
    <!-- Subject input -->
    <input
      v-model="subjectInput"
      type="text"
      placeholder='e.g. orders.> or >'
      :disabled="isRunning"
      class="font-mono text-sm px-2.5 py-1.5 rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-600 focus:outline-none focus:ring-1 focus:ring-emerald-500 dark:focus:ring-emerald-400 disabled:opacity-50 disabled:cursor-not-allowed w-64"
      @keydown="handleKeydown"
    />

    <!-- Start / Stop -->
    <button
      v-if="!isRunning"
      :disabled="subjectInput.trim().length === 0"
      class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md bg-emerald-600 text-white hover:bg-emerald-700 disabled:opacity-40 disabled:cursor-not-allowed"
      @click="handleStart"
    >
      <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
        <path d="M8 5v14l11-7z"/>
      </svg>
      Start
    </button>
    <button
      v-else
      class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md bg-red-500 text-white hover:bg-red-600"
      @click="$emit('stop')"
    >
      <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
        <path d="M6 6h12v12H6z"/>
      </svg>
      Stop
    </button>

    <!-- Pause / Resume -->
    <button
      :disabled="!isRunning"
      class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed"
      @click="isPaused ? $emit('resume') : $emit('pause')"
    >
      <svg v-if="!isPaused" class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
        <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
      </svg>
      <svg v-else class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
        <path d="M8 5v14l11-7z"/>
      </svg>
      {{ isPaused ? 'Resume' : 'Pause' }}
    </button>

    <!-- Clear -->
    <button
      :disabled="messageCount === 0"
      class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed"
      @click="$emit('clear')"
    >
      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
      </svg>
      Clear
    </button>

    <!-- Spacer -->
    <div class="flex-1" />

    <!-- Publish -->
    <button
      class="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
      title="Publish a message"
      @click="$emit('publish')"
    >
      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5"/>
      </svg>
      Publish
    </button>

    <!-- Buffer counter -->
    <span
      class="text-xs font-mono"
      :class="{
        'text-gray-500 dark:text-gray-400': bufferPercent < 80,
        'text-amber-600 dark:text-amber-400': bufferPercent >= 80 && bufferPercent < 100,
        'text-red-600 dark:text-red-400': bufferPercent >= 100,
      }"
    >
      {{ messageCount.toLocaleString() }} / {{ maxBuffer.toLocaleString() }}
    </span>

    <!-- WS status indicator -->
    <div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
      <span
        class="w-2 h-2 rounded-full shrink-0"
        :class="{
          'bg-green-500': status === 'open',
          'bg-yellow-400': status === 'connecting',
          'bg-red-500': status === 'closed' || status === 'error',
        }"
      />
      <span>
        {{ status === 'open' ? 'connected' : status === 'connecting' ? 'connecting' : status === 'error' ? 'error' : 'disconnected' }}
      </span>
    </div>
  </div>
</template>
