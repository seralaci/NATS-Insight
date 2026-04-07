<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useKvStore } from '../../stores/kv'

const emit = defineEmits<{
  close: []
  created: [name: string]
}>()

const kvStore = useKvStore()

const form = reactive({
  name: '',
  description: '',
  history: 1,
  ttl: '',
  limitMarkerTtl: '',
  replicas: 1,
  storage: 'file',
  compression: false,
  tags: [] as string[],
  tagInput: '',
  metadata: [] as { key: string; value: string }[],
  maxBucketSize: '',
  maxValueSize: '',
  republish: false,
  republishSrc: '',
  republishDest: '',
})

const saving = ref(false)
const serverError = ref('')
const showConfirm = ref(false)
const fieldErrors = reactive<Record<string, string>>({})

const inputClass = 'w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent'

// Parse duration string like "5m", "24h" to nanoseconds (Go duration)
function durationToNanos(s: string): number {
  if (!s) return 0
  const match = s.match(/^(\d+(?:\.\d+)?)\s*(ns|us|ms|s|m|h|d)$/)
  if (!match) return 0
  const val = parseFloat(match[1])
  const unit = match[2]
  const multipliers: Record<string, number> = {
    ns: 1, us: 1000, ms: 1e6, s: 1e9, m: 60e9, h: 3600e9, d: 86400e9
  }
  return val * (multipliers[unit] || 0)
}

// Parse size string like "100MB", "1GB" to bytes
function sizeToBytes(s: string): number {
  if (!s) return -1
  const match = s.match(/^(\d+(?:\.\d+)?)\s*(B|KB|MB|GB|TB)$/i)
  if (!match) return -1
  const val = parseFloat(match[1])
  const unit = match[2].toUpperCase()
  const multipliers: Record<string, number> = {
    B: 1, KB: 1024, MB: 1024**2, GB: 1024**3, TB: 1024**4
  }
  return val * (multipliers[unit] || 1)
}

// Config preview shows NATS stream config format (like Qaze)
const configPreview = computed(() => {
  const cfg: any = {
    allow_direct: true,
    allow_rollup_hdrs: true,
    compression: form.compression ? 's2' : 'none',
    deny_delete: true,
    deny_purge: false,
    description: form.description || '',
    discard: 'new',
    max_age: durationToNanos(form.ttl),
    max_bytes: form.maxBucketSize ? sizeToBytes(form.maxBucketSize) : -1,
    max_consumers: -1,
    max_msg_size: form.maxValueSize ? sizeToBytes(form.maxValueSize) : -1,
    max_msgs: -1,
    max_msgs_per_subject: form.history,
    metadata: {} as any,
    name: `KV_${form.name}`,
    num_replicas: form.replicas,
    retention: 'limits',
    storage: form.storage,
    subjects: [`$KV.${form.name}.>`],
  }
  if (form.metadata.length > 0) {
    for (const m of form.metadata.filter(m => m.key)) {
      cfg.metadata[m.key] = m.value
    }
  }
  if (form.limitMarkerTtl) {
    cfg.duplicate_window = durationToNanos(form.limitMarkerTtl)
  }
  if (form.republish && form.republishSrc && form.republishDest) {
    cfg.republish = { src: form.republishSrc, dest: form.republishDest }
  }
  return JSON.stringify(cfg, null, 2)
})

const configLines = computed(() => configPreview.value.split('\n'))

function addTag() {
  const t = form.tagInput.trim()
  if (t && !form.tags.includes(t)) {
    form.tags.push(t)
  }
  form.tagInput = ''
}

function removeTag(tag: string) {
  form.tags = form.tags.filter(t => t !== tag)
}

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
    fieldErrors.name = 'Bucket name is required'
  } else if (!/^[A-Za-z0-9_-]+$/.test(form.name)) {
    fieldErrors.name = 'Bucket name may only contain letters, numbers, underscores, and hyphens'
  }
  return Object.keys(fieldErrors).length === 0
}

function handleShowConfig() {
  if (!validate()) return
  showConfirm.value = true
}

async function submit() {
  if (!validate()) return
  saving.value = true
  serverError.value = ''
  try {
    const payload: any = { name: form.name }
    if (form.description) payload.description = form.description
    payload.history = form.history
    payload.replicas = form.replicas
    payload.storage = form.storage
    payload.compression = form.compression
    if (form.ttl) payload.ttl = form.ttl
    if (form.limitMarkerTtl) payload.limitMarkerTtl = form.limitMarkerTtl
    if (form.tags.length > 0) payload.tags = form.tags
    if (form.metadata.length > 0) {
      payload.metadata = Object.fromEntries(form.metadata.filter(m => m.key).map(m => [m.key, m.value]))
    }
    if (form.maxBucketSize) payload.maxBucketSize = form.maxBucketSize
    if (form.maxValueSize) payload.maxValueSize = form.maxValueSize
    if (form.republish) {
      payload.republish = { src: form.republishSrc, dest: form.republishDest }
    }
    await kvStore.createBucket(payload)
    emit('created', form.name)
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
    <!-- Confirm Config Modal (overlay on top of create modal) -->
    <div v-if="showConfirm" class="fixed inset-0 z-[60] flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="showConfirm = false"></div>
      <div class="relative w-full max-w-2xl max-h-[80vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Confirm KV Bucket Creation</h2>
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
              <!-- Line numbers -->
              <div class="py-3 px-3 text-right select-none text-gray-500 border-r border-gray-700 shrink-0">
                <div v-for="(_, i) in configLines" :key="i">{{ i + 1 }}</div>
              </div>
              <!-- Code -->
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
          >{{ saving ? 'Creating...' : 'Create KV Bucket' }}</button>
        </div>
      </div>
    </div>

    <!-- Main Create Modal -->
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="$emit('close')"></div>

      <div class="relative w-full max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Create KV Bucket</h2>
          <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300" @click="$emit('close')">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Error -->
        <div v-if="serverError" class="mx-6 mt-4 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800 shrink-0">
          {{ serverError }}
        </div>

        <!-- Body -->
        <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">

          <!-- Basic Configuration (full width) -->
          <section>
            <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Basic Configuration</h3>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Bucket Name <span class="text-red-500">*</span></label>
                <input
                  v-model="form.name"
                  type="text"
                  :class="[inputClass, fieldErrors.name ? 'border-red-400 dark:border-red-600' : '']"
                  placeholder="my-bucket"
                  @input="delete (fieldErrors as any).name"
                />
                <p v-if="fieldErrors.name" class="mt-1 text-xs text-red-500">{{ fieldErrors.name }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
                <input v-model="form.description" type="text" :class="inputClass" placeholder="Optional description" />
              </div>
            </div>
            <!-- Metadata (under basic) -->
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

          <!-- Value Retention (left) + Infrastructure (right) — side by side like Qaze -->
          <div class="grid grid-cols-2 gap-6">
            <!-- Value Retention -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Value Retention</h3>
              <div class="space-y-3">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">History Values</label>
                  <div class="flex items-center gap-2">
                    <button type="button" class="w-7 h-7 flex items-center justify-center rounded border border-gray-300 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 text-sm" @click="form.history = Math.max(1, form.history - 1)">-</button>
                    <input v-model.number="form.history" type="number" min="1" max="64" class="w-16 px-2 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 text-center" />
                    <button type="button" class="w-7 h-7 flex items-center justify-center rounded border border-gray-300 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 text-sm" @click="form.history = Math.min(64, form.history + 1)">+</button>
                  </div>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Maximum Age</label>
                  <input v-model="form.ttl" type="text" :class="inputClass" placeholder="e.g. 5m, 24h, 3600s" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Limit Marker TTL</label>
                  <input v-model="form.limitMarkerTtl" type="text" :class="inputClass" placeholder="e.g. 5m, 24h" />
                </div>
              </div>
            </section>

            <!-- Infrastructure -->
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
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Compression</label>
                  <button type="button" class="relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none" :class="form.compression ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-700'" @click="form.compression = !form.compression">
                    <span class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow" :class="form.compression ? 'translate-x-4' : 'translate-x-0.5'"></span>
                  </button>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Tags</label>
                  <input v-model="form.tagInput" type="text" :class="inputClass" placeholder="Press space or enter to add" @keydown.enter.prevent="addTag" @keydown.space.prevent="addTag" />
                  <div v-if="form.tags.length > 0" class="flex flex-wrap gap-1 mt-2">
                    <span v-for="tag in form.tags" :key="tag" class="flex items-center gap-1 px-2 py-0.5 text-xs bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300 rounded-full">
                      {{ tag }}
                      <button type="button" class="hover:text-red-500" @click="removeTag(tag)">
                        <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/></svg>
                      </button>
                    </span>
                  </div>
                </div>
              </div>
            </section>
          </div>

          <!-- Bucket Limits (left) + Republish (right) — side by side like Qaze -->
          <div class="grid grid-cols-2 gap-6">
            <!-- Bucket Limits -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Bucket Limits</h3>
              <div class="space-y-3">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Bucket Size</label>
                  <input v-model="form.maxBucketSize" type="text" :class="inputClass" placeholder="e.g. 100MB, 1GB" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Value Size</label>
                  <input v-model="form.maxValueSize" type="text" :class="inputClass" placeholder="e.g. 1MB, 256KB" />
                </div>
              </div>
            </section>

            <!-- Republish Configuration -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Republish Configuration</h3>
              <div class="flex items-center gap-3 mb-3">
                <span class="text-sm text-gray-700 dark:text-gray-300">Enable Republish</span>
                <button type="button" class="relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none" :class="form.republish ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-700'" @click="form.republish = !form.republish">
                  <span class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow" :class="form.republish ? 'translate-x-4' : 'translate-x-0.5'"></span>
                </button>
              </div>
              <div v-if="form.republish" class="space-y-3">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Source</label>
                  <input v-model="form.republishSrc" type="text" :class="inputClass" placeholder="Source subject" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Destination</label>
                  <input v-model="form.republishDest" type="text" :class="inputClass" placeholder="Destination subject" />
                </div>
              </div>
            </section>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0 flex items-center gap-3">
          <button type="button" class="flex items-center gap-1.5 px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="handleShowConfig">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5"/></svg>
            Show Config
          </button>
          <div class="flex-1"></div>
          <button type="button" class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="$emit('close')">Cancel</button>
          <button type="button" class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50" :disabled="saving" @click="submit">{{ saving ? 'Creating...' : 'Create KV Bucket' }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
