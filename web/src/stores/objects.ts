import { defineStore } from 'pinia'
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { objectsApi, type ObjectStoreInfo, type ObjectInfo } from '../lib/api'
import { useConnectionsStore } from './connections'

export const useObjectsStore = defineStore('objects', () => {
  const stores = ref<ObjectStoreInfo[]>([])
  const storesLoading = ref(false)
  const storesError = ref('')

  const selectedStore = ref<ObjectStoreInfo | null>(null)
  const objects = ref<ObjectInfo[]>([])
  const objectsLoading = ref(false)
  const objectsError = ref('')

  const actionLoading = ref(false)
  const actionError = ref('')

  async function fetchStores() {
    const connStore = useConnectionsStore()
    if (!connStore.status.connected) return
    storesLoading.value = true
    storesError.value = ''
    try {
      stores.value = await objectsApi.listStores()
    } catch (e: any) {
      storesError.value = e.message
      stores.value = []
    } finally {
      storesLoading.value = false
    }
  }

  async function createStore(data: any): Promise<ObjectStoreInfo> {
    actionLoading.value = true
    actionError.value = ''
    try {
      const store = await objectsApi.createStore(data)
      await fetchStores()
      toast.success('Object store created')
      return store
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function deleteStore(name: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await objectsApi.deleteStore(name)
      stores.value = stores.value.filter(s => s.bucket !== name)
      if (selectedStore.value?.bucket === name) {
        selectedStore.value = null
        objects.value = []
      }
      toast.success('Object store deleted')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  async function selectStore(name: string) {
    objects.value = []
    try {
      selectedStore.value = await objectsApi.getStore(name)
    } catch (e: any) {
      selectedStore.value = null
    }
    await fetchObjects(name)
  }

  async function fetchObjects(storeName: string) {
    objectsLoading.value = true
    objectsError.value = ''
    try {
      objects.value = await objectsApi.listObjects(storeName)
    } catch (e: any) {
      objectsError.value = e.message
      objects.value = []
    } finally {
      objectsLoading.value = false
    }
  }

  async function deleteObject(storeName: string, objectName: string) {
    actionLoading.value = true
    actionError.value = ''
    try {
      await objectsApi.deleteObject(storeName, objectName)
      objects.value = objects.value.filter(o => o.name !== objectName)
      toast.success('Object deleted')
    } catch (e: any) {
      actionError.value = e.message
      throw e
    } finally {
      actionLoading.value = false
    }
  }

  function clear() {
    stores.value = []
    storesError.value = ''
    selectedStore.value = null
    objects.value = []
    objectsError.value = ''
    actionError.value = ''
  }

  return {
    stores, storesLoading, storesError,
    selectedStore, objects, objectsLoading, objectsError,
    actionLoading, actionError,
    fetchStores, createStore, deleteStore,
    selectStore, fetchObjects, deleteObject,
    clear,
  }
})
