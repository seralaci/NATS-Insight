<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { useObjectsStore } from '../../stores/objects'

const props = defineProps<{
  storeName: string
}>()

const objectsStore = useObjectsStore()

const uploading = ref(false)
const uploadError = ref('')
const confirmDelete = ref<string | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`
}

function formatDate(iso: string): string {
  if (!iso) return '-'
  const d = new Date(iso)
  return d.toLocaleString()
}

async function handleUpload(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploading.value = true
  uploadError.value = ''
  const formData = new FormData()
  formData.append('file', file)
  try {
    await fetch(`/api/v1/objects/stores/${props.storeName}/objects`, {
      method: 'POST',
      body: formData,
    })
    await objectsStore.fetchObjects(props.storeName)
    toast.success('Object uploaded')
  } catch (e: any) {
    uploadError.value = e.message
  } finally {
    uploading.value = false
    if (fileInputRef.value) fileInputRef.value.value = ''
  }
}

async function download(name: string) {
  const resp = await fetch(`/api/v1/objects/stores/${props.storeName}/objects/${encodeURIComponent(name)}/data`)
  const blob = await resp.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = name
  a.click()
  URL.revokeObjectURL(url)
}

async function handleDelete(name: string) {
  if (confirmDelete.value !== name) {
    confirmDelete.value = name
    return
  }
  confirmDelete.value = null
  await objectsStore.deleteObject(props.storeName, name)
}

function cancelDelete() {
  confirmDelete.value = null
}
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="flex items-center justify-between px-4 py-2.5 border-b border-gray-200 dark:border-gray-800 shrink-0">
      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ objectsStore.objects.length }} object{{ objectsStore.objects.length !== 1 ? 's' : '' }}
      </span>
      <div class="flex items-center gap-2">
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700 disabled:opacity-50"
          :disabled="uploading"
          @click="fileInputRef?.click()"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5"/>
          </svg>
          {{ uploading ? 'Uploading...' : 'Upload' }}
        </button>
        <input
          ref="fileInputRef"
          type="file"
          class="hidden"
          @change="handleUpload"
        />
        <button
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          title="Refresh"
          @click="objectsStore.fetchObjects(storeName)"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"/>
          </svg>
        </button>
      </div>
    </div>

    <div v-if="uploadError" class="mx-4 mt-3 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800 shrink-0">
      {{ uploadError }}
    </div>

    <div class="flex-1 overflow-auto">
      <div v-if="objectsStore.objectsLoading" class="flex items-center justify-center py-12 text-sm text-gray-400 dark:text-gray-600">
        Loading...
      </div>
      <div v-else-if="objectsStore.objectsError" class="px-4 py-4 text-sm text-red-500">{{ objectsStore.objectsError }}</div>
      <div v-else-if="objectsStore.objects.length === 0" class="flex flex-col items-center justify-center py-20 gap-4">
        <svg class="w-12 h-12 text-gray-300 dark:text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
          <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z"/>
        </svg>
        <p class="text-sm text-gray-500 dark:text-gray-400">No objects found</p>
        <button
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 text-white rounded-md hover:bg-emerald-700"
          @click="fileInputRef?.click()"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5"/>
          </svg>
          Upload your first object
        </button>
      </div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-800">
            <th class="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Name</th>
            <th class="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Size</th>
            <th class="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Chunks</th>
            <th class="px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Modified</th>
            <th class="px-4 py-2.5 text-right text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="obj in objectsStore.objects"
            :key="obj.name"
            class="border-b border-gray-100 dark:border-gray-800/50 hover:bg-gray-50 dark:hover:bg-gray-800/40"
          >
            <td class="px-4 py-2.5 text-gray-800 dark:text-gray-200 font-mono text-xs truncate max-w-xs">{{ obj.name }}</td>
            <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400 whitespace-nowrap">{{ formatBytes(obj.size) }}</td>
            <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ obj.chunks }}</td>
            <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400 whitespace-nowrap text-xs">{{ formatDate(obj.modified) }}</td>
            <td class="px-4 py-2.5 text-right">
              <div class="flex items-center justify-end gap-1">
                <button
                  class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-emerald-600 dark:hover:text-emerald-400"
                  title="Download"
                  @click="download(obj.name)"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3"/>
                  </svg>
                </button>
                <button
                  class="p-1 rounded text-xs"
                  :class="confirmDelete === obj.name
                    ? 'bg-red-600 text-white hover:bg-red-700 px-2 py-1 rounded-md'
                    : 'hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-red-500'"
                  :title="confirmDelete === obj.name ? 'Click to confirm delete' : 'Delete'"
                  @click="handleDelete(obj.name)"
                  @blur="cancelDelete"
                >
                  <span v-if="confirmDelete === obj.name" class="text-xs">Confirm</span>
                  <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
