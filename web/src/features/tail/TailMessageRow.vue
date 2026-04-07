<script setup lang="ts">
import { ref, computed } from 'vue'
import type { TailMessage } from '../../composables/useTail'

const props = defineProps<{
  message: TailMessage
}>()

const expanded = ref(false)
const copied = ref(false)

const timestamp = computed(() => {
  const d = new Date(props.message.receivedAt)
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  const ms = String(d.getMilliseconds()).padStart(3, '0')
  return `${hh}:${mm}:${ss}.${ms}`
})

const parsedPayload = computed<{ isJson: boolean; display: string }>(() => {
  const text = props.message.dataText ?? ''
  if (!text) return { isJson: false, display: '' }
  try {
    const obj = JSON.parse(text)
    return { isJson: true, display: JSON.stringify(obj, null, 2) }
  } catch {
    return { isJson: false, display: text }
  }
})

const headerEntries = computed(() => {
  const h = props.message.headers
  if (!h) return []
  return Object.entries(h).flatMap(([key, values]) =>
    values.map((v) => ({ key, value: v }))
  )
})

const showExpander = computed(() => {
  const lines = parsedPayload.value.display.split('\n').length
  return lines > 4
})

async function copyPayload() {
  try {
    await navigator.clipboard.writeText(parsedPayload.value.display)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1500)
  } catch {
    // clipboard not available
  }
}
</script>

<template>
  <div
    class="group rounded-md border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 text-sm overflow-hidden"
  >
    <!-- Header row -->
    <div class="flex items-center justify-between px-3 py-1.5 bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
      <span class="font-mono font-semibold text-gray-900 dark:text-gray-100 text-xs">
        {{ message.subject }}
      </span>
      <div class="flex items-center gap-2">
        <!-- Copy button (visible on group hover) -->
        <button
          class="opacity-0 group-hover:opacity-100 flex items-center gap-1 px-1.5 py-0.5 text-xs rounded text-gray-500 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700"
          :title="copied ? 'Copied!' : 'Copy payload'"
          @click="copyPayload"
        >
          <svg v-if="!copied" class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184"/>
          </svg>
          <svg v-else class="w-3 h-3 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5"/>
          </svg>
        </button>
        <span class="font-mono text-xs text-gray-400 dark:text-gray-500">{{ timestamp }}</span>
      </div>
    </div>

    <!-- Headers section -->
    <div v-if="headerEntries.length > 0" class="px-3 py-1.5 border-b border-gray-200 dark:border-gray-800">
      <div class="flex items-center gap-1 mb-1">
        <span class="text-[10px] font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-600">Headers</span>
      </div>
      <div class="space-y-0.5">
        <div v-for="(entry, i) in headerEntries" :key="i" class="flex gap-2 font-mono text-xs">
          <span class="text-gray-500 dark:text-gray-400 shrink-0">{{ entry.key }}:</span>
          <span class="text-gray-700 dark:text-gray-300 break-all">{{ entry.value }}</span>
        </div>
      </div>
    </div>

    <!-- Payload -->
    <div class="relative">
      <pre
        class="px-3 py-2 font-mono text-xs text-gray-800 dark:text-gray-200 whitespace-pre-wrap break-all overflow-hidden leading-relaxed"
        :class="{ 'max-h-[5.5rem]': !expanded && showExpander }"
      >{{ parsedPayload.display || '(empty)' }}</pre>

      <!-- Expand / collapse -->
      <div v-if="showExpander" class="px-3 pb-1.5">
        <button
          class="text-xs text-emerald-600 dark:text-emerald-400 hover:underline"
          @click="expanded = !expanded"
        >
          {{ expanded ? 'Show less' : 'Show more' }}
        </button>
      </div>
    </div>
  </div>
</template>
