<script setup lang="ts">
import { ref, watch, reactive } from 'vue'
import { useConnectionsStore } from '../../stores/connections'
import type { Connection } from '../../lib/api'

const props = defineProps<{
  connection?: Connection | null
}>()

const emit = defineEmits<{
  close: []
  saved: [Connection]
}>()

const connStore = useConnectionsStore()

const form = ref({
  name: '',
  url: 'nats://localhost:4222',
  authMethod: 'none',
  username: '',
  password: '',
  token: '',
  nkey: '',
  credsFile: '',
  monitorUrl: 'http://localhost:8222',
})

const saving = ref(false)
const serverError = ref('')
const confirmDelete = ref(false)
const isEditing = ref(false)
const fieldErrors = reactive<Record<string, string>>({})
const monitorUrlManuallyEdited = ref(false)

// Auto-derive monitor URL from NATS URL
function deriveMonitorUrl(natsUrl: string): string {
  try {
    const match = natsUrl.match(/^nats:\/\/([^:\/]+)(?::(\d+))?/)
    if (match) {
      const host = match[1]
      return `http://${host}:8222`
    }
  } catch {}
  return 'http://localhost:8222'
}

function onUrlChange() {
  if (!monitorUrlManuallyEdited.value) {
    form.value.monitorUrl = deriveMonitorUrl(form.value.url)
  }
  delete fieldErrors.url
}

function onMonitorUrlChange() {
  monitorUrlManuallyEdited.value = true
}

watch(() => props.connection, (conn) => {
  if (conn) {
    isEditing.value = true
    form.value = {
      name: conn.name,
      url: conn.url,
      authMethod: conn.authMethod,
      username: conn.username || '',
      password: conn.password || '',
      token: conn.token || '',
      nkey: conn.nkey || '',
      credsFile: conn.credsFile || '',
      monitorUrl: conn.monitorUrl || 'http://localhost:8222',
    }
  } else {
    isEditing.value = false
    resetForm()
  }
  monitorUrlManuallyEdited.value = !!conn
}, { immediate: true })

function resetForm() {
  form.value = {
    name: '',
    url: 'nats://localhost:4222',
    authMethod: 'none',
    username: '',
    password: '',
    token: '',
    nkey: '',
    credsFile: '',
    monitorUrl: 'http://localhost:8222',
  }
  serverError.value = ''
  confirmDelete.value = false
  clearErrors()
}

function clearErrors() {
  Object.keys(fieldErrors).forEach(k => delete fieldErrors[k])
  serverError.value = ''
}

function validate(): boolean {
  clearErrors()
  if (!form.value.name.trim()) {
    fieldErrors.name = 'Connection name is required'
  }
  if (!form.value.url.trim()) {
    fieldErrors.url = 'URL is required'
  }
  return Object.keys(fieldErrors).length === 0
}

async function save() {
  if (!validate()) return
  saving.value = true
  serverError.value = ''
  try {
    let result: Connection
    if (isEditing.value && props.connection) {
      result = await connStore.updateConnection(props.connection.id, form.value)
    } else {
      result = await connStore.createConnection(form.value)
    }
    emit('saved', result)
  } catch (e: any) {
    serverError.value = e.message
  } finally {
    saving.value = false
  }
}

async function handleConnect() {
  if (!props.connection) return
  try {
    await connStore.connect(props.connection.id)
    emit('close')
  } catch (e: any) {
    serverError.value = e.message
  }
}

async function handleDelete() {
  if (!confirmDelete.value) {
    confirmDelete.value = true
    return
  }
  if (!props.connection) return
  try {
    await connStore.deleteConnection(props.connection.id)
    emit('close')
  } catch (e: any) {
    serverError.value = e.message
  }
}

const inputClass = 'w-full px-3 py-2 text-sm border rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent'
const inputNormal = 'border-gray-300 dark:border-gray-700'
const inputError = 'border-red-400 dark:border-red-600'
</script>

<template>
  <!-- Modal backdrop -->
  <Teleport to="body">
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <!-- Overlay -->
      <div class="absolute inset-0 bg-black/50" @click="$emit('close')"></div>

      <!-- Modal -->
      <div class="relative w-full max-w-2xl max-h-[85vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
            {{ isEditing ? connection?.name : 'Create Connection' }}
          </h2>
          <div class="flex items-center gap-2">
            <!-- Actions (edit mode) -->
            <template v-if="isEditing">
              <button
                class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-md bg-emerald-600 text-white hover:bg-emerald-700"
                @click="handleConnect"
              >
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z"/>
                </svg>
                Connect
              </button>
              <button
                class="flex items-center gap-1 px-3 py-1.5 text-sm rounded-md"
                :class="confirmDelete ? 'bg-red-600 text-white hover:bg-red-700' : 'text-red-500 hover:bg-red-50 dark:hover:bg-red-950'"
                @click="handleDelete"
              >
                {{ confirmDelete ? 'Click to confirm' : 'Delete' }}
              </button>
            </template>
            <button
              class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              @click="$emit('close')"
            >
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Server error (always visible at top) -->
        <div v-if="serverError" class="mx-6 mt-4 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800 shrink-0">
          {{ serverError }}
        </div>

        <!-- Form -->
        <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          <!-- Connection Name (full width) -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Connection Name</label>
            <input
              v-model="form.name"
              type="text"
              :class="[inputClass, fieldErrors.name ? inputError : inputNormal]"
              placeholder="My NATS Server"
              @input="delete fieldErrors.name"
            />
            <p v-if="fieldErrors.name" class="mt-1 text-xs text-red-500">{{ fieldErrors.name }}</p>
          </div>

          <!-- URL + Monitor URL (2 columns) -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">URL</label>
              <input
                v-model="form.url"
                type="text"
                :class="[inputClass, fieldErrors.url ? inputError : inputNormal]"
                placeholder="nats://localhost:4222"
                @input="onUrlChange"
              />
              <p v-if="fieldErrors.url" class="mt-1 text-xs text-red-500">{{ fieldErrors.url }}</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Monitor URL</label>
              <input
                v-model="form.monitorUrl"
                type="text"
                :class="[inputClass, inputNormal]"
                placeholder="http://localhost:8222"
                @input="onMonitorUrlChange"
              />
            </div>
          </div>

          <!-- Authentication Method + auth fields side by side -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Authentication Method</label>
              <div class="space-y-1.5">
                <label v-for="method in [
                  { value: 'none', label: 'None' },
                  { value: 'username_password', label: 'Username/Password' },
                  { value: 'token', label: 'Token' },
                ]" :key="method.value" class="flex items-center gap-2 cursor-pointer">
                  <input
                    v-model="form.authMethod"
                    type="radio"
                    :value="method.value"
                    class="text-emerald-500 focus:ring-emerald-500"
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ method.label }}</span>
                </label>
              </div>
            </div>

            <!-- Auth-specific fields (right column) -->
            <div class="space-y-3" v-if="form.authMethod !== 'none'">
              <template v-if="form.authMethod === 'username_password'">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username</label>
                  <input v-model="form.username" type="text" :class="[inputClass, inputNormal]" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Password</label>
                  <input v-model="form.password" type="password" :class="[inputClass, inputNormal]" />
                </div>
              </template>

              <div v-if="form.authMethod === 'token'">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Token</label>
                <input v-model="form.token" type="password" :class="[inputClass, inputNormal]" />
              </div>

            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0">
          <button
            class="w-full py-2.5 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="saving"
            @click="save"
          >
            {{ saving ? 'Saving...' : (isEditing ? 'Save Connection' : 'Create Connection') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
