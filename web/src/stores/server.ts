import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { serverApi } from '../lib/api'
import { useConnectionsStore } from './connections'
import { useKvStore } from './kv'
import { useStreamsStore } from './streams'
import { useObjectsStore } from './objects'

export const useServerStore = defineStore('server', () => {
  const jsz = ref<any>(null)
  const varz = ref<any>(null)
  const loading = ref(false)
  const error = ref('')

  const streamCount = computed(() => {
    const streamsStore = useStreamsStore()
    return streamsStore.streams.length || jsz.value?.streams || 0
  })
  const consumerCount = computed(() => jsz.value?.consumers ?? 0)
  const kvBucketCount = computed(() => {
    const kvStore = useKvStore()
    return kvStore.buckets.length
  })
  const objectBucketCount = computed(() => {
    const objectsStore = useObjectsStore()
    return objectsStore.stores.length
  })

  async function fetchData() {
    const connStore = useConnectionsStore()
    if (!connStore.status.connected) return
    loading.value = true
    error.value = ''
    try {
      const [jszData, varzData] = await Promise.all([
        serverApi.jsz().catch(() => null),
        serverApi.varz().catch(() => null),
      ])
      jsz.value = jszData
      varz.value = varzData
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  function clear() {
    jsz.value = null
    varz.value = null
    error.value = ''
  }

  return { jsz, varz, loading, error, streamCount, consumerCount, kvBucketCount, objectBucketCount, fetchData, clear }
})
