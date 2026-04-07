<script setup lang="ts">
import { ref, reactive } from 'vue'
import { toast } from 'vue-sonner'
import { publishApi } from '../../lib/api'

const props = defineProps<{
  initialSubject?: string
}>()

const emit = defineEmits<{
  close: []
  published: []
}>()

const inputClass =
  'w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent'

const form = reactive({
  subject: props.initialSubject || '',
  data: '',
  headers: [] as { key: string; value: string }[],
  mode: 'publish' as 'publish' | 'request',
  timeout: '5s',
})

const sending = ref(false)
const serverError = ref('')
const subjectError = ref('')
const replyData = ref<{ subject: string; data: string; dataText?: string; headers?: Record<string, string[]> } | null>(null)

function addHeader() {
  form.headers.push({ key: '', value: '' })
}

function removeHeader(i: number) {
  form.headers.splice(i, 1)
}

function formatJson() {
  try {
    const parsed = JSON.parse(form.data)
    form.data = JSON.stringify(parsed, null, 2)
  } catch {
    // not valid JSON, do nothing
  }
}

function buildHeaders(): Record<string, string[]> | undefined {
  const headers = form.headers.filter(h => h.key).reduce((acc, h) => {
    if (!acc[h.key]) acc[h.key] = []
    acc[h.key].push(h.value)
    return acc
  }, {} as Record<string, string[]>)
  return Object.keys(headers).length > 0 ? headers : undefined
}

async function submit() {
  subjectError.value = ''
  serverError.value = ''
  replyData.value = null

  if (!form.subject.trim()) {
    subjectError.value = 'Subject is required'
    return
  }

  sending.value = true
  try {
    const headers = buildHeaders()

    if (form.mode === 'publish') {
      await publishApi.publish({
        subject: form.subject,
        data: form.data || undefined,
        headers,
      })
      toast.success('Message published')
      emit('published')
      emit('close')
    } else {
      const resp = await publishApi.request({
        subject: form.subject,
        data: form.data || undefined,
        headers,
        timeout: form.timeout,
      })
      replyData.value = resp
    }
  } catch (e: any) {
    serverError.value = e.message
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="$emit('close')"></div>

      <div class="relative w-full max-w-2xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Publish Message</h2>
          <button
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            @click="$emit('close')"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Error -->
        <div
          v-if="serverError"
          class="mx-6 mt-4 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800 shrink-0"
        >
          {{ serverError }}
        </div>

        <!-- Body -->
        <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">

          <!-- Subject -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              Subject <span class="text-red-500">*</span>
            </label>
            <input
              v-model="form.subject"
              type="text"
              placeholder="e.g. orders.new"
              :class="[inputClass, subjectError ? 'border-red-500 dark:border-red-500 focus:ring-red-500' : '']"
              class="font-mono"
            />
            <p v-if="subjectError" class="mt-1 text-xs text-red-500">{{ subjectError }}</p>
          </div>

          <!-- Mode toggle -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Mode</label>
            <div class="flex gap-4">
              <label class="flex items-center gap-2 cursor-pointer select-none">
                <input
                  v-model="form.mode"
                  type="radio"
                  value="publish"
                  class="text-emerald-600 focus:ring-emerald-500 border-gray-300 dark:border-gray-700"
                />
                <span class="text-sm text-gray-700 dark:text-gray-300">Publish</span>
              </label>
              <label class="flex items-center gap-2 cursor-pointer select-none">
                <input
                  v-model="form.mode"
                  type="radio"
                  value="request"
                  class="text-emerald-600 focus:ring-emerald-500 border-gray-300 dark:border-gray-700"
                />
                <span class="text-sm text-gray-700 dark:text-gray-300">Request-Reply</span>
              </label>
            </div>
          </div>

          <!-- Timeout (request mode only) -->
          <div v-if="form.mode === 'request'">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Timeout</label>
            <input
              v-model="form.timeout"
              type="text"
              placeholder="e.g. 5s"
              :class="inputClass"
              class="w-32"
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Go duration string (e.g. 5s, 2m)</p>
          </div>

          <!-- Payload -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Payload</label>
              <button
                type="button"
                class="text-xs text-emerald-600 dark:text-emerald-400 hover:underline"
                @click="formatJson"
              >
                Format JSON
              </button>
            </div>
            <textarea
              v-model="form.data"
              rows="10"
              placeholder='{"key": "value"}'
              :class="inputClass"
              class="font-mono resize-y"
            ></textarea>
          </div>

          <!-- Headers -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Headers</label>
              <button
                type="button"
                class="text-xs text-emerald-600 dark:text-emerald-400 hover:underline"
                @click="addHeader"
              >
                + Add Header
              </button>
            </div>
            <div v-if="form.headers.length > 0" class="space-y-2">
              <div v-for="(h, i) in form.headers" :key="i" class="flex gap-2 items-center">
                <input
                  v-model="h.key"
                  type="text"
                  placeholder="Key"
                  :class="inputClass"
                  class="font-mono"
                />
                <input
                  v-model="h.value"
                  type="text"
                  placeholder="Value"
                  :class="inputClass"
                  class="font-mono"
                />
                <button
                  type="button"
                  class="text-gray-400 hover:text-red-500 shrink-0"
                  @click="removeHeader(i)"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
            </div>
            <p v-else class="text-xs text-gray-400 dark:text-gray-600 italic">No headers</p>
          </div>

          <!-- Reply panel (request-reply response) -->
          <div v-if="replyData" class="rounded-lg border border-emerald-300 dark:border-emerald-700 bg-emerald-50 dark:bg-emerald-950/30 p-4">
            <div class="flex items-center gap-2 mb-2">
              <svg class="w-4 h-4 text-emerald-600 dark:text-emerald-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/>
              </svg>
              <span class="text-sm font-medium text-emerald-700 dark:text-emerald-300">Reply received</span>
              <span class="text-xs font-mono text-emerald-600 dark:text-emerald-400 ml-auto">{{ replyData.subject }}</span>
            </div>
            <div v-if="replyData.headers && Object.keys(replyData.headers).length > 0" class="mb-2 space-y-0.5">
              <div v-for="(vals, key) in replyData.headers" :key="key" class="text-xs font-mono text-emerald-600 dark:text-emerald-400">
                <span class="text-gray-500 dark:text-gray-400">{{ key }}:</span> {{ vals.join(', ') }}
              </div>
            </div>
            <pre class="text-xs font-mono text-gray-800 dark:text-gray-200 whitespace-pre-wrap break-all bg-white dark:bg-gray-900 rounded p-3 border border-gray-200 dark:border-gray-800 max-h-48 overflow-y-auto">{{ replyData.dataText || replyData.data }}</pre>
          </div>

        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0 flex items-center gap-3">
          <div class="flex-1"></div>
          <button
            type="button"
            class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            @click="$emit('close')"
          >
            Cancel
          </button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50"
            :disabled="sending"
            @click="submit"
          >
            <span v-if="sending">{{ form.mode === 'request' ? 'Sending...' : 'Publishing...' }}</span>
            <span v-else>{{ form.mode === 'request' ? 'Send Request' : 'Publish' }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
