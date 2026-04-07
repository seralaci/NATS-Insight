import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { toast } from 'vue-sonner'
import { connectionsApi, type Connection, type ConnectionStatus } from '../lib/api'

export const useConnectionsStore = defineStore('connections', () => {
  const connections = ref<Connection[]>([])
  const status = ref<ConnectionStatus>({ connected: false })
  const loading = ref(false)
  const error = ref<string | null>(null)

  const activeConnection = computed(() => {
    if (!status.value.connectionId) return null
    return connections.value.find(c => c.id === status.value.connectionId) ?? null
  })

  async function fetchConnections() {
    try {
      connections.value = await connectionsApi.list()
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function fetchStatus() {
    try {
      status.value = await connectionsApi.status()
    } catch (e: any) {
      status.value = { connected: false }
    }
  }

  async function createConnection(data: Partial<Connection>) {
    loading.value = true
    error.value = null
    try {
      const conn = await connectionsApi.create(data)
      connections.value.push(conn)
      toast.success('Connection created')
      return conn
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function updateConnection(id: string, data: Partial<Connection>) {
    loading.value = true
    error.value = null
    try {
      const updated = await connectionsApi.update(id, data)
      const idx = connections.value.findIndex(c => c.id === id)
      if (idx >= 0) connections.value[idx] = updated
      toast.success('Connection updated')
      return updated
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function deleteConnection(id: string) {
    loading.value = true
    error.value = null
    try {
      await connectionsApi.delete(id)
      connections.value = connections.value.filter(c => c.id !== id)
      if (status.value.connectionId === id) {
        status.value = { connected: false }
      }
      toast.success('Connection deleted')
    } catch (e: any) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  async function connect(id: string) {
    loading.value = true
    error.value = null
    try {
      status.value = await connectionsApi.connect(id)
      toast.success('Connected')
    } catch (e: any) {
      error.value = e.message
      status.value = { connected: false, connectionId: id, error: e.message }
      throw e
    } finally {
      loading.value = false
    }
  }

  async function disconnect() {
    loading.value = true
    try {
      if (status.value.connectionId) {
        status.value = await connectionsApi.disconnect(status.value.connectionId)
        toast.success('Disconnected')
      }
    } catch (e: any) {
      status.value = { connected: false }
    } finally {
      loading.value = false
    }
  }

  return {
    connections, status, loading, error, activeConnection,
    fetchConnections, fetchStatus, createConnection, updateConnection,
    deleteConnection, connect, disconnect,
  }
})
