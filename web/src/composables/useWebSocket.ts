import { ref, onUnmounted } from 'vue'
import { createWsClient, type WsStatus, type WsClient } from '../lib/ws'

export function useWebSocket(url: string, onMessage: (data: any) => void) {
  const status = ref<WsStatus>('closed')
  let client: WsClient | null = null

  function connect() {
    client = createWsClient({
      url,
      onMessage,
      onStatusChange: (s) => {
        status.value = s
      },
    })
  }

  function disconnect() {
    client?.close()
    client = null
    status.value = 'closed'
  }

  function send(data: any) {
    client?.send(data)
  }

  onUnmounted(() => disconnect())

  return { status, connect, disconnect, send }
}
