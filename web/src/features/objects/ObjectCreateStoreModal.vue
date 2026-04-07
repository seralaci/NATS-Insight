<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useObjectsStore } from '../../stores/objects'

const emit = defineEmits<{
  close: []
  created: [name: string]
}>()

const objectsStore = useObjectsStore()

const form = reactive({
  name: '',
  description: '',
  maxChunkSize: '',
  maxBytes: '',
  storage: 'file',
  replicas: 1,
  metadata: [] as { key: string; value: string }[],
})

const saving = ref(false)
const serverError = ref('')
const showConfirm = ref(false)
const createAnother = ref(false)
const fieldErrors = reactive<Record<string, string>>({})

const inputClass = 'w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent'

function sizeToBytes(s: string): number {
  if (!s) return -1
  const match = s.match(/^(\d+(?:\.\d+)?)\s*(B|KB|MB|GB|TB)$/i)
  if (!match) return -1
  const val = parseFloat(match[1])
  const unit = match[2].toUpperCase()
  const multipliers: Record<string, number> = {
    B: 1, KB: 1024, MB: 1024 ** 2, GB: 1024 ** 3, TB: 1024 ** 4,
  }
  return val * (multipliers[unit] || 1)
}

const configPreview = computed(() => {
  const cfg: any = {
    bucket: form.name,
    description: form.description || '',
    storage: form.storage,
    num_replicas: form.replicas,
  }
  if (form.maxChunkSize) cfg.max_chunk_size = sizeToBytes(form.maxChunkSize)
  if (form.maxBytes) cfg.max_bytes = sizeToBytes(form.maxBytes)
  if (form.metadata.length > 0) {
    cfg.metadata = {}
    for (const m of form.metadata.filter(m => m.key)) {
      cfg.metadata[m.key] = m.value
    }
  }
  return JSON.stringify(cfg, null, 2)
})

const configLines = computed(() => configPreview.value.split('\n'))

function addMetadata() {
  form.metadata.push({ key: '', value: '' })
}

function removeMetadata(i: number) {
  form.metadata.splice(i, 1)
}

function validate(): boolean {
  Object.keys(fieldErrors).forEach(k => delete (fieldErrors as any)[k])
  serverError.value = ''
  if (!form.name.trim()) {
    fieldErrors.name = 'Store name is required'
  } else if (!/^[A-Za-z0-9_-]+$/.test(form.name)) {
    fieldErrors.name = 'Store name may only contain letters, numbers, underscores, and hyphens'
  }
  return Object.keys(fieldErrors).length === 0
}

function handleShowConfig() {
  if (!validate()) return
  showConfirm.value = true
}

function resetForm() {
  form.name = ''
  form.description = ''
  form.maxChunkSize = ''
  form.maxBytes = ''
  form.storage = 'file'
  form.replicas = 1
  form.metadata = []
  serverError.value = ''
  showConfirm.value = false
}

async function submit() {
  if (!validate()) return
  saving.value = true
  serverError.value = ''
  try {
    const payload: any = { name: form.name }
    if (form.description) payload.description = form.description
    payload.storage = form.storage
    payload.replicas = form.replicas
    if (form.maxChunkSize) payload.maxChunkSize = form.maxChunkSize
    if (form.maxBytes) payload.maxBytes = form.maxBytes
    if (form.metadata.length > 0) {
      payload.metadata = Object.fromEntries(form.metadata.filter(m => m.key).map(m => [m.key, m.value]))
    }
    await objectsStore.createStore(payload)
    const created = form.name
    if (createAnother.value) {
      resetForm()
    } else {
      emit('created', created)
    }
  } catch (e: any) {
    serverError.value = e.message
    showConfirm.value = false
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div v-if="showConfirm" class="fixed inset-0 z-[60] flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="showConfirm = false"></div>
      <div class="relative w-full max-w-2xl max-h-[80vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Confirm Object Store Creation</h2>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">Please review the configuration that will be created.</p>
          </div>
          <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400" @click="showConfirm = false">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <div class="flex-1 overflow-y-auto p-4">
          <div class="bg-gray-900 dark:bg-gray-900 rounded-lg overflow-hidden border border-gray-700">
            <div class="flex text-xs font-mono">
              <div class="py-3 px-3 text-right select-none text-gray-500 border-r border-gray-700 shrink-0">
                <div v-for="(_, i) in configLines" :key="i">{{ i + 1 }}</div>
              </div>
              <div class="py-3 px-4 overflow-x-auto text-green-400 whitespace-pre">{{ configPreview }}</div>
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0 flex items-center gap-3">
          <div class="flex-1"></div>
          <button
            type="button"
            class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            @click="showConfirm = false"
          >Cancel</button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50"
            :disabled="saving"
            @click="submit"
          >{{ saving ? 'Creating...' : 'Create Object Store' }}</button>
        </div>
      </div>
    </div>

    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="$emit('close')"></div>

      <div class="relative w-full max-w-2xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Create Object Store</h2>
          <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300" @click="$emit('close')">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <div v-if="serverError" class="mx-6 mt-4 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800 shrink-0">
          {{ serverError }}
        </div>

        <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">
          <section>
            <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Basic Configuration</h3>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Store Name <span class="text-red-500">*</span></label>
                <input
                  v-model="form.name"
                  type="text"
                  :class="[inputClass, fieldErrors.name ? 'border-red-400 dark:border-red-600' : '']"
                  placeholder="my-store"
                  @input="delete (fieldErrors as any).name"
                />
                <p v-if="fieldErrors.name" class="mt-1 text-xs text-red-500">{{ fieldErrors.name }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
                <input v-model="form.description" type="text" :class="inputClass" placeholder="Optional description" />
              </div>
            </div>
            <div class="mt-3">
              <div class="flex items-center justify-between mb-1">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Metadata</label>
                <button type="button" class="text-xs text-emerald-600 dark:text-emerald-400 hover:underline" @click="addMetadata">+ Add Metadata</button>
              </div>
              <div v-if="form.metadata.length > 0" class="space-y-2">
                <div v-for="(m, i) in form.metadata" :key="i" class="flex gap-2">
                  <input v-model="m.key" type="text" :class="inputClass" placeholder="Key" />
                  <input v-model="m.value" type="text" :class="inputClass" placeholder="Value" />
                  <button type="button" class="text-gray-400 hover:text-red-500 shrink-0" @click="removeMetadata(i)">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </section>

          <div class="grid grid-cols-2 gap-6">
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Limits</h3>
              <div class="space-y-3">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Chunk Size</label>
                  <input v-model="form.maxChunkSize" type="text" :class="inputClass" placeholder="e.g. 256KB, 1MB" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Bytes</label>
                  <input v-model="form.maxBytes" type="text" :class="inputClass" placeholder="e.g. 1GB, 500MB" />
                </div>
              </div>
            </section>

            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Infrastructure</h3>
              <div class="space-y-3">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Replicas</label>
                  <div class="flex items-center gap-2">
                    <button type="button" class="w-7 h-7 flex items-center justify-center rounded border border-gray-300 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 text-sm" @click="form.replicas = Math.max(1, form.replicas - 1)">-</button>
                    <input v-model.number="form.replicas" type="number" min="1" max="5" class="w-16 px-2 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 text-center" />
                    <button type="button" class="w-7 h-7 flex items-center justify-center rounded border border-gray-300 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 text-sm" @click="form.replicas = Math.min(5, form.replicas + 1)">+</button>
                  </div>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Storage</label>
                  <div class="flex items-center gap-4">
                    <label class="flex items-center gap-2 cursor-pointer">
                      <input v-model="form.storage" type="radio" value="file" class="text-emerald-500 focus:ring-emerald-500" />
                      <span class="text-sm text-gray-700 dark:text-gray-300">File</span>
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer">
                      <input v-model="form.storage" type="radio" value="memory" class="text-emerald-500 focus:ring-emerald-500" />
                      <span class="text-sm text-gray-700 dark:text-gray-300">Memory</span>
                    </label>
                  </div>
                </div>
              </div>
            </section>
          </div>
        </div>

        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0 flex items-center gap-3">
          <button type="button" class="flex items-center gap-1.5 px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="handleShowConfig">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5"/>
            </svg>
            Show Config
          </button>
          <label class="flex items-center gap-2 cursor-pointer ml-2">
            <input v-model="createAnother" type="checkbox" class="text-emerald-500 focus:ring-emerald-500 rounded" />
            <span class="text-sm text-gray-600 dark:text-gray-400">Create another</span>
          </label>
          <div class="flex-1"></div>
          <button type="button" class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="$emit('close')">Cancel</button>
          <button type="button" class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50" :disabled="saving" @click="submit">{{ saving ? 'Creating...' : 'Create Object Store' }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
