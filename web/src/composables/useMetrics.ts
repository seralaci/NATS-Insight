import { ref, shallowRef, onUnmounted } from 'vue'
import { createWsClient, type WsClient, type WsStatus } from '../lib/ws'

export interface MetricPoint {
  timestamp: number
  msgsIn: number
  msgsOut: number
  bytesIn: number
  bytesOut: number
  connections: number
}

export function useMetrics(maxPoints = 60) {
  const points = shallowRef<MetricPoint[]>([])
  const connected = ref(false)
  const status = ref<WsStatus>('closed')

  let client: WsClient | null = null
  let prevVarz: any = null

  function handleMessage(data: any) {
    if (data.type === 'metrics' && data.varz) {
      const varz = data.varz
      if (prevVarz) {
        const point: MetricPoint = {
          timestamp: Date.now(),
          msgsIn: Math.max(0, (varz.in_msgs ?? 0) - (prevVarz.in_msgs ?? 0)),
          msgsOut: Math.max(0, (varz.out_msgs ?? 0) - (prevVarz.out_msgs ?? 0)),
          bytesIn: Math.max(0, (varz.in_bytes ?? 0) - (prevVarz.in_bytes ?? 0)),
          bytesOut: Math.max(0, (varz.out_bytes ?? 0) - (prevVarz.out_bytes ?? 0)),
          connections: varz.connections ?? 0,
        }
        const newPoints = [...points.value, point]
        if (newPoints.length > maxPoints) newPoints.shift()
        points.value = newPoints
      }
      prevVarz = varz
    }

    if (data.type === 'stopped') {
      connected.value = false
    }
  }

  function start(interval = 5) {
    stop()
    prevVarz = null
    points.value = []

    client = createWsClient({
      url: '/api/v1/ws/metrics',
      onMessage: handleMessage,
      onStatusChange: (s) => {
        status.value = s
        connected.value = s === 'open'
        if (s === 'open') {
          client!.send({ type: 'start', interval })
        }
      },
      reconnect: false,
    })
  }

  function stop() {
    if (client) {
      try {
        client.send({ type: 'stop' })
      } catch {}
      client.close()
      client = null
    }
    connected.value = false
    status.value = 'closed'
  }

  onUnmounted(stop)

  return { points, connected, status, start, stop }
}
