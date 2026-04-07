import { defineStore } from 'pinia'
import { ref } from 'vue'
import { streamsApi, consumersApi, type StreamInfo, type StreamMessage, type ConsumerInfo } from '../lib/api'
import { useConnectionsStore } from './connections'
import { toast } from 'vue-sonner'

export const useStreamsStore = defineStore('streams', () => {
  const streams = ref<StreamInfo[]>([])
  const streamsLoading = ref(false)
  const streamsError = ref('')

  const selectedStream = ref<StreamInfo | null>(null)

  // Messages
  const messages = ref<StreamMessage[]>([])
  const messagesLoading = ref(false)
  const messagesFirstSeq = ref(0)
  const messagesLastSeq = ref(0)
  const messagesTotal = ref(0)

  // Consumers
  const consumers = ref<ConsumerInfo[]>([])
  const consumersLoading = ref(false)
  const selectedConsumer = ref<ConsumerInfo | null>(null)

  const actionLoading = ref(false)
  const actionError = ref('')

  async function fetchStreams() {
    const connStore = useConnectionsStore()
    if (!connStore.status.connected) return
    streamsLoading.value = true
    streamsError.value = ''
    try {
      streams.value = await streamsApi.list()
    } catch (e: any) {
      streamsError.value = e.message
      streams.value = []
    } finally {
      streamsLoading.value = false
    }
  }

  async function selectStream(name: string) {
    // Only clear data when switching to a different stream
    const switching = selectedStream.value?.config.name !== name
    if (switching) {
      messages.value = []
      consumers.value = []
      selectedConsumer.value = null
    }
    try {
      selectedStream.value = await streamsApi.get(name)
      if (!switching) {
        // Re-fetch visible data on refresh
        if (consumers.value.length > 0) {
          await fetchConsumers(name)
        }
        if (messages.value.length > 0) {
          await fetchMessages(name, { limit: 50 })
        }
      }
    } catch (e: any) {
      selectedStream.value = null
    }
  }

  async function createStream(data: any): Promise<StreamInfo> {
    actionLoading.value = true
    actionError.value = ''
    try {
      const stream = await streamsApi.create(data)
      await fetchStreams()
      toast.success('Stream created')
      return stream
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function updateStream(name: string, data: any): Promise<StreamInfo> {
    actionLoading.value = true
    actionError.value = ''
    try {
      const stream = await streamsApi.update(name, data)
      await fetchStreams()
      if (selectedStream.value?.config.name === name) {
        selectedStream.value = stream
      }
      toast.success('Stream updated')
      return stream
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function deleteStream(name: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await streamsApi.delete(name)
      streams.value = streams.value.filter(s => s.config.name !== name)
      if (selectedStream.value?.config.name === name) {
        selectedStream.value = null
        messages.value = []
        consumers.value = []
        selectedConsumer.value = null
      }
      toast.success('Stream deleted')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function purgeStream(name: string, opts?: { subject?: string; seq?: number; keep?: number }) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await streamsApi.purge(name, opts)
      // Refresh stream info after purge
      if (selectedStream.value?.config.name === name) {
        selectedStream.value = await streamsApi.get(name)
      }
      toast.success('Stream purged')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function fetchMessages(streamName: string, opts?: { startSeq?: number; limit?: number; subject?: string; startDate?: string }) {
    messagesLoading.value = true
    try {
      const result = await streamsApi.listMessages(streamName, opts)
      messages.value = result.messages ?? []
      messagesFirstSeq.value = result.firstSeq
      messagesLastSeq.value = result.lastSeq
      messagesTotal.value = result.total
    } catch (e: any) {
      messages.value = []
    } finally {
      messagesLoading.value = false
    }
  }

  async function fetchMoreMessages(streamName: string, opts?: { startSeq?: number; limit?: number; subject?: string }) {
    messagesLoading.value = true
    try {
      const result = await streamsApi.listMessages(streamName, opts)
      const newMsgs = result.messages ?? []
      // Append, dedup by sequence
      const existingSeqs = new Set(messages.value.map(m => m.sequence))
      for (const m of newMsgs) {
        if (!existingSeqs.has(m.sequence)) {
          messages.value.push(m)
        }
      }
      messagesFirstSeq.value = result.firstSeq
      messagesLastSeq.value = result.lastSeq
      messagesTotal.value = result.total
    } catch (e: any) {
      // ignore
    } finally {
      messagesLoading.value = false
    }
  }

  async function fetchConsumers(streamName: string) {
    consumersLoading.value = true
    try {
      consumers.value = await consumersApi.list(streamName)
    } catch (e: any) {
      consumers.value = []
    } finally {
      consumersLoading.value = false
    }
  }

  async function deleteConsumer(streamName: string, consumerName: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await consumersApi.delete(streamName, consumerName)
      consumers.value = consumers.value.filter(c => c.name !== consumerName)
      if (selectedConsumer.value?.name === consumerName) {
        selectedConsumer.value = null
      }
      toast.success('Consumer deleted')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function pauseConsumer(streamName: string, consumerName: string, pauseUntil: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await consumersApi.pause(streamName, consumerName, pauseUntil)
      await fetchConsumers(streamName)
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function resumeConsumer(streamName: string, consumerName: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await consumersApi.resume(streamName, consumerName)
      await fetchConsumers(streamName)
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  function clear() {
    streams.value = []
    streamsError.value = ''
    selectedStream.value = null
    messages.value = []
    consumers.value = []
    selectedConsumer.value = null
    actionError.value = ''
  }

  return {
    streams, streamsLoading, streamsError,
    selectedStream,
    messages, messagesLoading, messagesFirstSeq, messagesLastSeq, messagesTotal,
    consumers, consumersLoading, selectedConsumer,
    actionLoading, actionError,
    fetchStreams, selectStream,
    createStream, updateStream, deleteStream, purgeStream,
    fetchMessages, fetchMoreMessages,
    fetchConsumers, deleteConsumer, pauseConsumer, resumeConsumer,
    clear,
  }
})
