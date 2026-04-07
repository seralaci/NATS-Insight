import { defineStore } from 'pinia'
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { kvApi, type KvBucket, type KvEntry } from '../lib/api'
import { useConnectionsStore } from './connections'

export const useKvStore = defineStore('kv', () => {
  const buckets = ref<KvBucket[]>([])
  const bucketsLoading = ref(false)
  const bucketsError = ref('')

  const selectedBucket = ref<KvBucket | null>(null)
  const keys = ref<string[]>([])
  const keysLoading = ref(false)

  const selectedKey = ref<string | null>(null)
  const keyDetail = ref<KvEntry | null>(null)
  const keyHistory = ref<KvEntry[]>([])
  const keyDetailLoading = ref(false)

  const actionLoading = ref(false)
  const actionError = ref('')

  async function fetchBuckets() {
    const connStore = useConnectionsStore()
    if (!connStore.status.connected) return
    bucketsLoading.value = true
    bucketsError.value = ''
    try {
      buckets.value = await kvApi.listBuckets()
    } catch (e: any) {
      bucketsError.value = e.message
      buckets.value = []
    } finally {
      bucketsLoading.value = false
    }
  }

  async function createBucket(data: any): Promise<KvBucket> {
    actionLoading.value = true
    actionError.value = ''
    try {
      const bucket = await kvApi.createBucket(data)
      await fetchBuckets()
      toast.success('Bucket created')
      return bucket
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function deleteBucket(name: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await kvApi.deleteBucket(name)
      buckets.value = buckets.value.filter(b => b.name !== name)
      if (selectedBucket.value?.name === name) {
        selectedBucket.value = null
        keys.value = []
        selectedKey.value = null
        keyDetail.value = null
        keyHistory.value = []
      }
      toast.success('Bucket deleted')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function fetchKeys(bucketName: string, q?: string) {
    keysLoading.value = true
    try {
      keys.value = await kvApi.listKeys(bucketName, q)
    } catch (e: any) {
      keys.value = []
    } finally {
      keysLoading.value = false
    }
  }

  async function selectBucket(bucketName: string) {
    selectedKey.value = null
    keyDetail.value = null
    keyHistory.value = []
    try {
      selectedBucket.value = await kvApi.getBucket(bucketName)
    } catch (e: any) {
      selectedBucket.value = null
    }
    await fetchKeys(bucketName)
  }

  async function fetchKeyDetail(bucketName: string, key: string) {
    selectedKey.value = key
    keyDetailLoading.value = true
    keyDetail.value = null
    keyHistory.value = []
    try {
      const [entry, history] = await Promise.all([
        kvApi.getKey(bucketName, key),
        kvApi.getKeyHistory(bucketName, key).catch(() => [] as KvEntry[]),
      ])
      keyDetail.value = entry
      keyHistory.value = history
    } catch (e: any) {
      keyDetail.value = null
    } finally {
      keyDetailLoading.value = false
    }
  }

  async function putKey(bucketName: string, key: string, value: string): Promise<KvEntry> {
    actionLoading.value = true
    actionError.value = ''
    try {
      const entry = await kvApi.putKey(bucketName, key, value)
      await fetchKeys(bucketName)
      toast.success('Key saved')
      return entry
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function deleteKey(bucketName: string, key: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await kvApi.deleteKey(bucketName, key)
      keys.value = keys.value.filter(k => k !== key)
      if (selectedKey.value === key) {
        selectedKey.value = null
        keyDetail.value = null
        keyHistory.value = []
      }
      toast.success('Key deleted')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function purgeKey(bucketName: string, key: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await kvApi.purgeKey(bucketName, key)
      keys.value = keys.value.filter(k => k !== key)
      if (selectedKey.value === key) {
        selectedKey.value = null
        keyDetail.value = null
        keyHistory.value = []
      }
      toast.success('Key purged')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  function clear() {
    buckets.value = []
    bucketsError.value = ''
    selectedBucket.value = null
    keys.value = []
    selectedKey.value = null
    keyDetail.value = null
    keyHistory.value = []
    actionError.value = ''
  }

  return {
    buckets, bucketsLoading, bucketsError,
    selectedBucket, keys, keysLoading,
    selectedKey, keyDetail, keyHistory, keyDetailLoading,
    actionLoading, actionError,
    fetchBuckets, createBucket, deleteBucket,
    fetchKeys, selectBucket,
    fetchKeyDetail, putKey, deleteKey, purgeKey,
    clear,
  }
})
