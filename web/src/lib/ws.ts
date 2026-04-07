export type WsStatus = 'connecting' | 'open' | 'closed' | 'error'

export interface WsClientOptions {
  url: string
  onMessage: (data: any) => void
  onStatusChange?: (status: WsStatus) => void
  reconnect?: boolean
  reconnectDelay?: number
  maxReconnectDelay?: number
}

export interface WsClient {
  send(data: any): void
  close(): void
  readonly status: WsStatus
}

export function createWsClient(options: WsClientOptions): WsClient {
  const {
    onMessage,
    onStatusChange,
  } = options

  let shouldReconnect = options.reconnect !== false
  const baseDelay = options.reconnectDelay ?? 2000
  const maxDelay = options.maxReconnectDelay ?? 30000

  let socket: WebSocket | null = null
  let currentStatus: WsStatus = 'connecting'
  let pendingMessages: string[] = []
  let reconnectAttempt = 0
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  function setStatus(s: WsStatus) {
    currentStatus = s
    onStatusChange?.(s)
  }

  function resolveUrl(): string {
    const wsUrl = new URL(options.url, window.location.href)
    wsUrl.protocol = wsUrl.protocol === 'https:' ? 'wss:' : 'ws:'
    return wsUrl.toString()
  }

  function connect() {
    setStatus('connecting')
    socket = new WebSocket(resolveUrl())

    socket.onopen = () => {
      reconnectAttempt = 0
      setStatus('open')
      for (const msg of pendingMessages) {
        socket!.send(msg)
      }
      pendingMessages = []
    }

    socket.onmessage = (event: MessageEvent) => {
      try {
        const parsed = JSON.parse(event.data as string)
        onMessage(parsed)
      } catch {
        // non-JSON frame: ignore
      }
    }

    socket.onclose = () => {
      socket = null
      if (currentStatus !== 'error') {
        setStatus('closed')
      }
      scheduleReconnect()
    }

    socket.onerror = () => {
      setStatus('error')
      // onclose fires after onerror, reconnect is handled there
    }
  }

  function scheduleReconnect() {
    if (!shouldReconnect) return
    const delay = Math.min(baseDelay * 2 ** reconnectAttempt, maxDelay)
    reconnectAttempt++
    reconnectTimer = setTimeout(() => {
      if (shouldReconnect) {
        connect()
      }
    }, delay)
  }

  function send(data: any) {
    const serialized = JSON.stringify(data)
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.send(serialized)
    } else if (currentStatus === 'connecting') {
      pendingMessages.push(serialized)
    }
    // closed or error: drop the message
  }

  function close() {
    shouldReconnect = false
    if (reconnectTimer !== null) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    pendingMessages = []
    if (socket) {
      socket.onclose = null
      socket.onerror = null
      socket.close()
      socket = null
    }
    setStatus('closed')
  }

  connect()

  return {
    send,
    close,
    get status() {
      return currentStatus
    },
  }
}
