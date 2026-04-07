<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useKvStore } from '../../stores/kv'
import type { KvEntry } from '../../lib/api'

const props = defineProps<{
  bucketName: string
  editEntry?: KvEntry | null
}>()

const emit = defineEmits<{
  close: []
  saved: [key: string]
}>()

const kvStore = useKvStore()

const keyInput = ref('')
const valueInput = ref('')
const format = ref<'json' | 'raw'>('json')
const optimistic = ref(true)
const saving = ref(false)
const serverError = ref('')
const keyError = ref('')

const isEdit = computed(() => !!props.editEntry)
const title = computed(() => isEdit.value ? 'Create New Revision' : 'Create Key')
const submitLabel = computed(() => isEdit.value ? 'Create New Revision' : 'Create Key')

watch(() => props.editEntry, (entry) => {
  if (entry) {
    keyInput.value = entry.key
    valueInput.value = entry.valueText || ''
    try {
      const parsed = JSON.parse(valueInput.value)
      valueInput.value = JSON.stringify(parsed, null, 2)
      format.value = 'json'
    } catch {
      format.value = 'raw'
    }
  } else {
    keyInput.value = ''
    valueInput.value = ''
    format.value = 'json'
  }
}, { immediate: true })

function formatJson() {
  try {
    const parsed = JSON.parse(valueInput.value)
    valueInput.value = JSON.stringify(parsed, null, 2)
    format.value = 'json'
  } catch {
    // leave as-is
  }
}

async function submit() {
  keyError.value = ''
  serverError.value = ''

  if (!keyInput.value.trim()) {
    keyError.value = 'Key is required'
    return
  }

  saving.value = true
  try {
    const entry = await kvStore.putKey(props.bucketName, keyInput.value.trim(), valueInput.value)
    if (isEdit.value) {
      await kvStore.fetchKeyDetail(props.bucketName, keyInput.value.trim())
    }
    emit('saved', entry.key)
  } catch (e: any) {
    serverError.value = e.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="$emit('close')"></div>

      <div class="relative w-full max-w-xl max-h-[85vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">{{ title }}</h2>
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
        <div v-if="serverError" class="mx-6 mt-4 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800 shrink-0">
          {{ serverError }}
        </div>

        <!-- Body -->
        <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          <!-- Key -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Key</label>
            <input
              v-model="keyInput"
              type="text"
              :readonly="isEdit"
              class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
              :class="[
                isEdit ? 'bg-gray-50 dark:bg-gray-800 cursor-default' : '',
                keyError ? 'border-red-400 dark:border-red-600' : ''
              ]"
              placeholder="my-key"
              @input="keyError = ''"
            />
            <p v-if="keyError" class="mt-1 text-xs text-red-500">{{ keyError }}</p>
          </div>

          <!-- Value -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Value</label>
              <div class="flex items-center gap-2">
                <div class="flex rounded-md overflow-hidden border border-gray-300 dark:border-gray-700">
                  <button
                    type="button"
                    class="px-2.5 py-1 text-xs"
                    :class="format === 'json' ? 'bg-emerald-600 text-white' : 'bg-white dark:bg-gray-900 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
                    @click="format = 'json'"
                  >JSON</button>
                  <button
                    type="button"
                    class="px-2.5 py-1 text-xs border-l border-gray-300 dark:border-gray-700"
                    :class="format === 'raw' ? 'bg-emerald-600 text-white' : 'bg-white dark:bg-gray-900 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'"
                    @click="format = 'raw'"
                  >Raw</button>
                </div>
                <button
                  v-if="format === 'json'"
                  type="button"
                  class="px-2 py-1 text-xs text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 border border-gray-300 dark:border-gray-700 rounded"
                  @click="formatJson"
                  title="Format JSON"
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12a7.5 7.5 0 0 0 15 0m-15 0a7.5 7.5 0 1 1 15 0m-15 0H3m16.5 0H21m-1.5 0H12m-8.457 3.077 1.41-.513m14.095-5.13 1.41-.513M5.106 17.785l1.15-.964m11.49-9.642 1.149-.964M7.501 19.795l.75-1.3m7.5-12.99.75-1.3m-6.063 16.658.26-1.477m2.605-14.772.26-1.477m0 17.726-.26-1.477M10.698 4.614l-.26-1.477M16.5 19.794l-.75-1.299M7.5 4.205 12 12m6.894 5.785-1.149-.964M6.256 7.178l-1.15-.964m15.352 8.864-1.41-.513M4.954 9.435l-1.41-.514M12.002 12l-3.75 6.495"/>
                  </svg>
                </button>
              </div>
            </div>
            <textarea
              v-model="valueInput"
              rows="10"
              class="w-full px-3 py-2 text-sm font-mono border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent resize-y"
              :placeholder="format === 'json' ? '{ }' : 'Enter value...'"
            ></textarea>
          </div>

          <!-- Optimistic concurrency (edit mode) -->
          <div v-if="isEdit" class="flex items-start gap-2">
            <input
              id="optimistic"
              v-model="optimistic"
              type="checkbox"
              class="mt-0.5 text-emerald-500 focus:ring-emerald-500 rounded"
            />
            <label for="optimistic" class="text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
              Make sure that revision <span class="font-medium">{{ editEntry?.revision }}</span> is the latest
              <span class="text-gray-400 dark:text-gray-500">(optimistic concurrency control)</span>
            </label>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0 flex items-center justify-end gap-3">
          <button
            type="button"
            class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
            @click="$emit('close')"
          >
            Cancel
          </button>
          <button
            type="button"
            class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="saving"
            @click="submit"
          >
            {{ saving ? 'Saving...' : submitLabel }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
