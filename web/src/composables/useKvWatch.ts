import { ref, shallowRef } from 'vue'
import { createWsClient, type WsClient, type WsStatus } from '../lib/ws'

export interface KvWatchEntry {
  key: string
  value: string
  valueText?: string
  revision: number
  operation: string
  created: string
  size: number
  receivedAt: string
}

export function useKvWatch(maxBuffer = 500) {
  const entries = shallowRef<KvWatchEntry[]>([])
  const isWatching = ref(false)
  const status = ref<WsStatus>('closed')

  let client: WsClient | null = null

  function handleMessage(data: any) {
    if (data.type === 'watching') {
      isWatching.value = true
      return
    }

    if (data.type === 'stopped') {
      isWatching.value = false
      return
    }

    if (data.type === 'entry') {
      let valueText: string | undefined
      if (data.value) {
        try {
          valueText = atob(data.value)
        } catch {
          valueText = undefined
        }
      }

      const entry: KvWatchEntry = {
        key: data.key ?? '',
        value: data.value ?? '',
        valueText,
        revision: data.revision ?? 0,
        operation: data.operation ?? '',
        created: data.created ?? new Date().toISOString(),
        size: data.size ?? 0,
        receivedAt: new Date().toISOString(),
      }

      const next = [...entries.value, entry]
      entries.value = next.length > maxBuffer ? next.slice(-maxBuffer) : next
    }
  }

  function watch(bucket: string, keyFilter?: string) {
    // Tear down any existing connection first
    stop()

    entries.value = []
    isWatching.value = false

    client = createWsClient({
      url: `/api/v1/ws/kv/${encodeURIComponent(bucket)}/watch`,
      onMessage: handleMessage,
      onStatusChange: (s) => {
        status.value = s
      },
    })

    // Send the watch command once connected (queued if still connecting)
    client.send({ type: 'watch', bucket, keyFilter: keyFilter || '>' })
  }

  function stop() {
    if (client) {
      client.send({ type: 'stop' })
      client.close()
      client = null
    }
    isWatching.value = false
    status.value = 'closed'
  }

  function clear() {
    entries.value = []
  }

  return {
    entries,
    isWatching,
    status,
    watch,
    stop,
    clear,
  }
}
