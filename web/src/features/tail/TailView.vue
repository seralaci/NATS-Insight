<script setup lang="ts">
import { ref } from 'vue'
import { useTail } from '../../composables/useTail'
import TailToolbar from './TailToolbar.vue'
import TailMessageList from './TailMessageList.vue'
import PublishModal from '../publish/PublishModal.vue'

const MAX_BUFFER = 1000
const tail = useTail(MAX_BUFFER)

const hasStarted = ref(false)
const showPublishModal = ref(false)
const currentSubject = ref('')

function handleStart(subject: string) {
  hasStarted.value = true
  currentSubject.value = subject
  tail.start(subject)
}

function handleStop() {
  tail.stop()
}

function handlePause() {
  tail.pause()
}

function handleResume() {
  tail.resume()
}

function handleClear() {
  tail.clear()
}

function openPublish() {
  showPublishModal.value = true
}
</script>

<template>
  <div class="flex flex-col h-full min-h-0 bg-gray-50 dark:bg-gray-950">
    <!-- Page heading -->
    <div class="px-4 pt-4 pb-2 shrink-0">
      <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">Live Tail</h1>
    </div>

    <!-- Toolbar -->
    <TailToolbar
      :is-running="tail.isRunning.value"
      :is-paused="tail.isPaused.value"
      :message-count="tail.messageCount.value"
      :max-buffer="MAX_BUFFER"
      :status="tail.status.value"
      @start="handleStart"
      @stop="handleStop"
      @pause="handlePause"
      @resume="handleResume"
      @clear="handleClear"
      @publish="openPublish"
    />

    <!-- Paused banner -->
    <div
      v-if="tail.isPaused.value"
      class="shrink-0 px-4 py-2 bg-amber-50 dark:bg-amber-950/30 border-b border-amber-200 dark:border-amber-800 text-xs text-amber-700 dark:text-amber-400 flex items-center gap-2"
    >
      <svg class="w-3.5 h-3.5 shrink-0" fill="currentColor" viewBox="0 0 24 24">
        <path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/>
      </svg>
      Paused — incoming messages are buffered and will appear when you resume.
    </div>

    <!-- Empty state -->
    <div
      v-if="!hasStarted && tail.messages.value.length === 0"
      class="flex-1 flex flex-col items-center justify-center text-center p-8 min-h-0"
    >
      <div class="w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center mb-4">
        <svg class="w-6 h-6 text-gray-400 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M8.288 15.038a5.25 5.25 0 0 1 7.424 0M5.106 11.856c3.807-3.808 9.98-3.808 13.788 0M1.924 8.674c5.565-5.565 14.587-5.565 20.152 0M12.53 18.22l-.53.53-.53-.53a.75.75 0 0 1 1.06 0Z"/>
        </svg>
      </div>
      <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subscribe to a subject to see messages</p>
      <p class="text-xs text-gray-400 dark:text-gray-600">
        Enter a subject above (e.g. <span class="font-mono">orders.&gt;</span> or <span class="font-mono">&gt;</span> for all) and click Start.
      </p>
    </div>

    <!-- Message list -->
    <TailMessageList
      v-else
      :messages="tail.messages.value"
    />

    <!-- Publish modal -->
    <PublishModal
      v-if="showPublishModal"
      :initial-subject="currentSubject"
      @close="showPublishModal = false"
      @published="showPublishModal = false"
    />
  </div>
</template>
