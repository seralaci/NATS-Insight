import { ref, shallowRef, onUnmounted } from 'vue'
import { createWsClient, type WsClient, type WsStatus } from '../lib/ws'

export interface TailMessage {
  id: string
  subject: string
  reply?: string
  headers?: Record<string, string[]>
  data: string
  dataText?: string
  size: number
  receivedAt: string
}

export function useTail(maxBuffer = 1000) {
  const messages = shallowRef<TailMessage[]>([])
  const isRunning = ref(false)
  const isPaused = ref(false)
  const subject = ref('')
  const messageCount = ref(0)
  const status = ref<WsStatus>('closed')

  let client: WsClient | null = null
  let pendingBuffer: TailMessage[] = []

  function handleMessage(data: any) {
    if (data.type === 'started') {
      isRunning.value = true
      return
    }

    if (data.type === 'stopped') {
      isRunning.value = false
      return
    }

    if (data.type === 'message') {
      let dataText: string | undefined
      if (data.data) {
        try {
          dataText = atob(data.data)
        } catch {
          dataText = undefined
        }
      }

      const msg: TailMessage = {
        id: crypto.randomUUID(),
        subject: data.subject ?? '',
        reply: data.reply,
        headers: data.headers,
        data: data.data ?? '',
        dataText,
        size: data.size ?? 0,
        receivedAt: new Date().toISOString(),
      }

      messageCount.value++

      if (isPaused.value) {
        pendingBuffer.push(msg)
        if (pendingBuffer.length > maxBuffer) {
          pendingBuffer = pendingBuffer.slice(-maxBuffer)
        }
        return
      }

      const next = [...messages.value, msg]
      messages.value = next.length > maxBuffer ? next.slice(-maxBuffer) : next
    }
  }

  function start(subjectFilter: string) {
    // Tear down any existing connection
    stop()

    subject.value = subjectFilter
    isPaused.value = false
    pendingBuffer = []
    messages.value = []
    messageCount.value = 0
    isRunning.value = false

    // Connect with subject as query param (backend reads it from URL)
    client = createWsClient({
      url: `/api/v1/ws/tail?subject=${encodeURIComponent(subjectFilter)}`,
      onMessage: handleMessage,
      onStatusChange: (s) => {
        status.value = s
      },
      reconnect: false,
    })
  }

  function stop() {
    if (client) {
      client.close()
      client = null
    }
    isRunning.value = false
    status.value = 'closed'
  }

  function pause() {
    isPaused.value = true
  }

  function resume() {
    isPaused.value = false
    if (pendingBuffer.length === 0) return

    const combined = [...messages.value, ...pendingBuffer]
    messages.value = combined.length > maxBuffer ? combined.slice(-maxBuffer) : combined
    pendingBuffer = []
  }

  function clear() {
    messages.value = []
    pendingBuffer = []
    messageCount.value = 0
  }

  onUnmounted(() => stop())

  return {
    messages,
    isRunning,
    isPaused,
    subject,
    messageCount,
    status,
    start,
    stop,
    pause,
    resume,
    clear,
  }
}
