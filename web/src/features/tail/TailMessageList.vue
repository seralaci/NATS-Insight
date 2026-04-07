<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { TailMessage } from '../../composables/useTail'
import TailMessageRow from './TailMessageRow.vue'

const props = defineProps<{
  messages: TailMessage[]
}>()

const container = ref<HTMLElement | null>(null)
const autoScroll = ref(true)

function isNearBottom(): boolean {
  const el = container.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 80
}

function onScroll() {
  autoScroll.value = isNearBottom()
}

async function scrollToBottom() {
  await nextTick()
  const el = container.value
  if (el) {
    el.scrollTo({ top: el.scrollHeight })
  }
}

watch(
  () => props.messages,
  () => {
    if (autoScroll.value) {
      scrollToBottom()
    }
  },
  { flush: 'post' }
)
</script>

<template>
  <div
    ref="container"
    class="flex-1 overflow-y-auto p-4 space-y-2 min-h-0"
    @scroll="onScroll"
  >
    <TailMessageRow
      v-for="msg in messages"
      :key="msg.id"
      :message="msg"
    />
  </div>
</template>
