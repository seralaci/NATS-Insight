<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useStreamsStore } from '../../stores/streams'
import type { StreamInfo } from '../../lib/api'

const props = defineProps<{
  editStream?: StreamInfo
  mirrorStream?: StreamInfo
  duplicateStream?: StreamInfo
}>()

const emit = defineEmits<{
  close: []
  created: [name: string]
  updated: [name: string]
}>()

const streamsStore = useStreamsStore()

const isEditMode = computed(() => !!props.editStream)
const isMirrorMode = computed(() => !!props.mirrorStream)
const isDuplicateMode = computed(() => !!props.duplicateStream)

// Helper: nanoseconds to duration string for display
function nanosToDisplay(ns: number): string {
  if (!ns || ns <= 0) return ''
  if (ns % (86400e9) === 0) return `${ns / 86400e9}d`
  if (ns % (3600e9) === 0) return `${ns / 3600e9}h`
  if (ns % (60e9) === 0) return `${ns / 60e9}m`
  if (ns % (1e9) === 0) return `${ns / 1e9}s`
  if (ns % (1e6) === 0) return `${ns / 1e6}ms`
  return `${ns}ns`
}

function makeDefaultForm() {
  return {
    name: '',
    description: '',
    subjects: [] as string[],
    subjectInput: '',
    metadata: [] as { key: string; value: string }[],
    // Retention
    retention: 'limits',
    discard: 'old',
    discardNewPerSubject: false,
    // Infrastructure
    replicas: 1,
    storage: 'file',
    compression: 'none',
    persistMode: 'default',
    tags: [] as string[],
    tagInput: '',
    // Message Processing
    firstSeq: '',
    duplicates: '',
    subjectDeleteMarkerTtl: '',
    subjectTransformSrc: '',
    subjectTransformDest: '',
    noAck: false,
    allowPerMessageTtl: false,
    // Stream Limits
    maxAge: '',
    maxBytes: '',
    maxMsgs: '',
    maxMsgsPerSubject: '',
    maxMsgSize: '',
    maxConsumers: '',
    consumerInactiveThreshold: '',
    maxAckPending: '',
    // Advanced Features
    allowRollup: false,
    allowDirect: false,
    mirrorDirect: false,
    allowMsgCounter: false,
    allowMsgSchedules: false,
    allowAtomicPublish: false,
    denyDelete: false,
    denyPurge: false,
    // Republish
    republish: false,
    republishSrc: '',
    republishDest: '',
    sources: [] as Array<{
      name: string
      filterSubject: string
      optStartSeq: string
      optStartTime: string
      optStartTimeTime: string
      external: boolean
      apiPrefix: string
      deliverPrefix: string
      subjectTransforms: Array<{ src: string; dest: string }>
    }>,
    // Mirror
    mirrorEnabled: false,
    mirrorName: '',
    mirrorFilterSubject: '',
    mirrorStartSeq: '',
    mirrorStartTime: '',
    mirrorStartTimeTime: '',
    mirrorExternal: false,
    mirrorApiPrefix: '$JS.API',
    mirrorDeliverPrefix: '$JS.INBOX',
  }
}

const form = reactive(makeDefaultForm())

// Pre-fill form from editStream when provided
watch(() => props.editStream, (es) => {
  if (!es) return
  const cfg = es.config
  form.name = cfg.name ?? ''
  form.description = cfg.description ?? ''
  form.subjects = [...(cfg.subjects ?? [])]
  form.subjectInput = ''
  form.metadata = cfg.metadata ? Object.entries(cfg.metadata).map(([k, v]) => ({ key: k, value: v })) : []
  form.retention = cfg.retention ?? 'limits'
  form.discard = cfg.discard ?? 'old'
  form.discardNewPerSubject = cfg.discardNewPerSubject ?? false
  form.replicas = cfg.replicas ?? 1
  form.storage = cfg.storage ?? 'file'
  form.compression = cfg.compression ?? 'none'
  form.noAck = cfg.noAck ?? false
  form.allowRollup = cfg.allowRollup ?? false
  form.allowDirect = cfg.allowDirect ?? false
  form.denyDelete = cfg.denyDelete ?? false
  form.denyPurge = cfg.denyPurge ?? false
  // Limits — convert nanos/numbers to display strings
  form.maxAge = cfg.maxAge > 0 ? nanosToDisplay(cfg.maxAge) : ''
  form.maxBytes = cfg.maxBytes > 0 ? String(cfg.maxBytes) : ''
  form.maxMsgs = cfg.maxMsgs > 0 ? String(cfg.maxMsgs) : ''
  form.maxMsgSize = cfg.maxMsgSize > 0 ? String(cfg.maxMsgSize) : ''
  form.maxMsgsPerSubject = cfg.maxMsgsPerSubject > 0 ? String(cfg.maxMsgsPerSubject) : ''
  form.maxConsumers = cfg.maxConsumers > 0 ? String(cfg.maxConsumers) : ''
  form.duplicates = cfg.duplicates > 0 ? nanosToDisplay(cfg.duplicates) : ''
  form.firstSeq = es.state?.first_seq ? String(es.state.first_seq) : ''
  // Sources
  form.sources = (cfg as any).sources?.map((s: any) => ({
    name: s.name || '',
    filterSubject: s.filter_subject || '',
    optStartSeq: s.opt_start_seq ? String(s.opt_start_seq) : '',
    optStartTime: s.opt_start_time ? s.opt_start_time.split('T')[0] : '',
    optStartTimeTime: s.opt_start_time ? s.opt_start_time.split('T')[1]?.replace('Z', '').split('.')[0] || '' : '',
    external: !!s.external,
    apiPrefix: s.external?.api || '$JS.API',
    deliverPrefix: s.external?.deliver || '$JS.INBOX',
    subjectTransforms: s.subject_transforms?.map((t: any) => ({ src: t.src || '', dest: t.dest || '' })) || [],
  })) || []
  // Mirror
  if ((cfg as any).mirror) {
    const m = (cfg as any).mirror
    form.mirrorEnabled = true
    form.mirrorName = m.name || ''
    form.mirrorFilterSubject = m.filter_subject || ''
    form.mirrorStartSeq = m.opt_start_seq ? String(m.opt_start_seq) : ''
    form.mirrorStartTime = m.opt_start_time ? m.opt_start_time.split('T')[0] : ''
    form.mirrorStartTimeTime = m.opt_start_time ? m.opt_start_time.split('T')[1]?.replace('Z', '').split('.')[0] || '' : ''
    form.mirrorExternal = !!m.external
    form.mirrorApiPrefix = m.external?.api || '$JS.API'
    form.mirrorDeliverPrefix = m.external?.deliver || '$JS.INBOX'
  }
}, { immediate: true })

watch(() => props.mirrorStream, (ms) => {
  if (!ms) return
  const cfg = ms.config
  form.name = cfg.name + '_mirror'
  form.description = cfg.description || ''
  form.mirrorEnabled = true
  form.mirrorName = cfg.name
  form.mirrorDirect = true
  // Set defaults from original stream
  form.retention = cfg.retention ?? 'limits'
  form.storage = cfg.storage ?? 'file'
  form.replicas = cfg.replicas ?? 1
  form.compression = cfg.compression ?? 'none'
}, { immediate: true })

watch(() => props.duplicateStream, (ds) => {
  if (!ds) return
  const cfg = ds.config
  // Same field mapping as editStream watcher, but with different name
  form.name = cfg.name + '_copy'
  form.description = cfg.description ?? ''
  form.subjects = [...(cfg.subjects ?? [])]
  form.subjectInput = ''
  form.metadata = cfg.metadata ? Object.entries(cfg.metadata).map(([k, v]) => ({ key: k, value: v })) : []
  form.retention = cfg.retention ?? 'limits'
  form.discard = cfg.discard ?? 'old'
  form.discardNewPerSubject = cfg.discardNewPerSubject ?? false
  form.replicas = cfg.replicas ?? 1
  form.storage = cfg.storage ?? 'file'
  form.compression = cfg.compression ?? 'none'
  form.noAck = cfg.noAck ?? false
  form.allowRollup = cfg.allowRollup ?? false
  form.allowDirect = cfg.allowDirect ?? false
  form.denyDelete = cfg.denyDelete ?? false
  form.denyPurge = cfg.denyPurge ?? false
  form.maxAge = cfg.maxAge > 0 ? nanosToDisplay(cfg.maxAge) : ''
  form.maxBytes = cfg.maxBytes > 0 ? String(cfg.maxBytes) : ''
  form.maxMsgs = cfg.maxMsgs > 0 ? String(cfg.maxMsgs) : ''
  form.maxMsgSize = cfg.maxMsgSize > 0 ? String(cfg.maxMsgSize) : ''
  form.maxMsgsPerSubject = cfg.maxMsgsPerSubject > 0 ? String(cfg.maxMsgsPerSubject) : ''
  form.maxConsumers = cfg.maxConsumers > 0 ? String(cfg.maxConsumers) : ''
  form.duplicates = cfg.duplicates > 0 ? nanosToDisplay(cfg.duplicates) : ''
  // Sources
  form.sources = (cfg as any).sources?.map((s: any) => ({
    name: s.name || '',
    filterSubject: s.filter_subject || '',
    optStartSeq: s.opt_start_seq ? String(s.opt_start_seq) : '',
    optStartTime: s.opt_start_time ? s.opt_start_time.split('T')[0] : '',
    optStartTimeTime: s.opt_start_time ? s.opt_start_time.split('T')[1]?.replace('Z', '').split('.')[0] || '' : '',
    external: !!s.external,
    apiPrefix: s.external?.api || '$JS.API',
    deliverPrefix: s.external?.deliver || '$JS.INBOX',
    subjectTransforms: s.subject_transforms?.map((t: any) => ({ src: t.src || '', dest: t.dest || '' })) || [],
  })) || []
  // Mirror (if duplicating a mirror stream)
  if ((cfg as any).mirror) {
    const m = (cfg as any).mirror
    form.mirrorEnabled = true
    form.mirrorName = m.name || ''
    form.mirrorFilterSubject = m.filter_subject || ''
    form.mirrorStartSeq = m.opt_start_seq ? String(m.opt_start_seq) : ''
    form.mirrorStartTime = m.opt_start_time ? m.opt_start_time.split('T')[0] : ''
    form.mirrorStartTimeTime = m.opt_start_time ? m.opt_start_time.split('T')[1]?.replace('Z', '').split('.')[0] || '' : ''
    form.mirrorExternal = !!m.external
    form.mirrorApiPrefix = m.external?.api || '$JS.API'
    form.mirrorDeliverPrefix = m.external?.deliver || '$JS.INBOX'
  }
}, { immediate: true })

const saving = ref(false)
const serverError = ref('')
const showConfirm = ref(false)
const createAnother = ref(false)
const fieldErrors = reactive<Record<string, string>>({})

const inputClass = 'w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent'

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

function durationToGo(s: string): string {
  if (!s) return ''
  const match = s.match(/^(\d+(?:\.\d+)?)\s*d$/)
  if (match) return `${parseFloat(match[1]) * 24}h`
  return s
}

function parseByteSize(s: string): number {
  if (!s) return 0
  const trimmed = s.trim().toUpperCase()
  const suffixes: [string, number][] = [
    ['TB', 1024 ** 4], ['GB', 1024 ** 3], ['MB', 1024 ** 2], ['KB', 1024], ['B', 1],
  ]
  for (const [suffix, mult] of suffixes) {
    if (trimmed.endsWith(suffix)) {
      const num = parseFloat(trimmed.slice(0, -suffix.length).trim())
      return isNaN(num) ? 0 : Math.round(num * mult)
    }
  }
  const num = parseFloat(trimmed)
  return isNaN(num) ? 0 : Math.round(num)
}

function addSubject() {
  const t = form.subjectInput.trim()
  if (t && !form.subjects.includes(t)) {
    form.subjects.push(t)
    delete (fieldErrors as any).subjects
  }
  form.subjectInput = ''
}

function removeSubject(s: string) {
  form.subjects = form.subjects.filter(x => x !== s)
}

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

function makeEmptySource() {
  return {
    name: '',
    filterSubject: '',
    optStartSeq: '',
    optStartTime: '',
    optStartTimeTime: '',
    external: false,
    apiPrefix: '$JS.API',
    deliverPrefix: '$JS.INBOX',
    subjectTransforms: [] as Array<{ src: string; dest: string }>,
  }
}

function addSource() {
  form.sources.push(makeEmptySource())
}

function removeSource(index: number) {
  form.sources.splice(index, 1)
}

function addSourceTransform(sourceIndex: number) {
  form.sources[sourceIndex].subjectTransforms.push({ src: '', dest: '' })
}

function removeSourceTransform(sourceIndex: number, transformIndex: number) {
  form.sources[sourceIndex].subjectTransforms.splice(transformIndex, 1)
}

const configPreview = computed(() => {
  const cfg: any = {
    name: form.name,
    subjects: form.subjects.length > 0 ? form.subjects : undefined,
    retention: form.retention,
    storage: form.storage,
    num_replicas: form.replicas,
    discard: form.discard,
    discard_new_per_subject: form.discardNewPerSubject || undefined,
    compression: form.compression !== 'none' ? form.compression : 'none',
    no_ack: form.noAck || undefined,
    allow_rollup_hdrs: form.allowRollup || undefined,
    allow_direct: form.allowDirect || undefined,
    mirror_direct: false,
    allow_msg_counter: form.allowMsgCounter || undefined,
    allow_msg_ttl: form.allowPerMessageTtl || undefined,
    allow_atomic: form.allowAtomicPublish || undefined,
    allow_msg_schedules: form.allowMsgSchedules || undefined,
    deny_delete: form.denyDelete || undefined,
    deny_purge: form.denyPurge || undefined,
  }
  if (form.description) cfg.description = form.description
  if (form.maxMsgs) cfg.max_msgs = parseInt(form.maxMsgs)
  if (form.maxBytes) cfg.max_bytes = parseByteSize(form.maxBytes)
  if (form.maxAge) cfg.max_age = durationToNanos(form.maxAge)
  if (form.maxMsgSize) cfg.max_msg_size = parseByteSize(form.maxMsgSize)
  if (form.maxMsgsPerSubject) cfg.max_msgs_per_subject = parseInt(form.maxMsgsPerSubject)
  if (form.maxConsumers) cfg.max_consumers = parseInt(form.maxConsumers)
  if (form.duplicates) cfg.duplicate_window = durationToNanos(form.duplicates)
  if (form.firstSeq) cfg.first_seq = parseInt(form.firstSeq)
  if (form.consumerInactiveThreshold || form.maxAckPending) {
    const limits: any = {}
    if (form.consumerInactiveThreshold) limits.inactive_threshold = durationToNanos(form.consumerInactiveThreshold)
    if (form.maxAckPending) limits.max_ack_pending = parseInt(form.maxAckPending)
    cfg.consumer_limits = limits
  }
  if (form.subjectDeleteMarkerTtl) cfg.subject_delete_marker_ttl = durationToNanos(form.subjectDeleteMarkerTtl)
  if (form.subjectTransformSrc || form.subjectTransformDest) {
    cfg.subject_transform = { src: form.subjectTransformSrc, dest: form.subjectTransformDest }
  }
  if (form.metadata.length > 0) {
    cfg.metadata = Object.fromEntries(form.metadata.filter(m => m.key).map(m => [m.key, m.value]))
  }
  if (form.republish && form.republishSrc && form.republishDest) {
    cfg.republish = { src: form.republishSrc, dest: form.republishDest }
  }
  if (form.tags.length > 0) cfg.placement = { tags: form.tags }
  if (form.persistMode !== 'default') cfg.persist_mode = form.persistMode
  // strip undefined
  for (const k of Object.keys(cfg)) {
    if (cfg[k] === undefined) delete cfg[k]
  }
  return JSON.stringify(cfg, null, 2)
})

const configLines = computed(() => configPreview.value.split('\n'))

function validate(): boolean {
  // Auto-add any pending subject input
  if (form.subjectInput.trim()) {
    addSubject()
  }
  Object.keys(fieldErrors).forEach(k => delete (fieldErrors as any)[k])
  serverError.value = ''
  if (!form.name.trim()) {
    fieldErrors.name = 'Stream name is required'
  } else if (!/^[A-Za-z0-9_-]+$/.test(form.name)) {
    fieldErrors.name = 'Stream name may only contain letters, numbers, underscores, and hyphens'
  }
  if (form.subjects.length === 0 && !form.mirrorEnabled) {
    fieldErrors.subjects = 'At least one subject is required'
  }
  if (form.allowMsgCounter && form.discard === 'new') {
    fieldErrors.discard = 'Message Counter cannot be used with Discard New'
  }
  if (form.discardNewPerSubject && form.discard !== 'new') {
    fieldErrors.discard = 'Discard New Per Subject requires Discard New policy'
  }
  if (form.allowAtomicPublish && form.persistMode === 'async') {
    fieldErrors.persistMode = 'Async Persist Mode cannot be used with Atomic Publish'
  }
  if (form.allowRollup && form.denyPurge) {
    fieldErrors.denyPurge = 'Rollup Headers requires Purge permission (disable Deny Purge)'
  }
  for (const [i, s] of form.sources.entries()) {
    if (!s.name.trim()) {
      fieldErrors[`source_${i}_name`] = 'Source stream name is required'
    }
  }
  return Object.keys(fieldErrors).length === 0
}

function handleShowConfig() {
  if (!validate()) return
  showConfirm.value = true
}

function resetForm() {
  const defaults = makeDefaultForm()
  Object.assign(form, defaults)
}

async function submit() {
  if (!validate()) return
  saving.value = true
  serverError.value = ''
  try {
    const payload: any = {
      name: form.name,
      retention: form.retention,
      storage: form.storage,
      replicas: form.replicas,
      discard: form.discard,
    }
    if (form.description) payload.description = form.description
    // Mirror
    if (form.mirrorEnabled && form.mirrorName) {
      const mirror: any = { name: form.mirrorName }
      if (form.mirrorFilterSubject) mirror.filterSubject = form.mirrorFilterSubject
      if (form.mirrorStartSeq) mirror.optStartSeq = parseInt(form.mirrorStartSeq)
      if (form.mirrorStartTime) {
        const timePart = form.mirrorStartTimeTime || '00:00:00'
        mirror.optStartTime = `${form.mirrorStartTime}T${timePart}Z`
      }
      if (form.mirrorExternal) {
        mirror.external = {
          apiPrefix: form.mirrorApiPrefix,
          deliverPrefix: form.mirrorDeliverPrefix,
        }
      }
      payload.mirror = mirror
    }
    if (form.subjects.length > 0 && !form.mirrorEnabled) payload.subjects = form.subjects
    if (form.discardNewPerSubject) payload.discardNewPerSubject = true
    if (form.compression !== 'none') payload.compression = form.compression
    if (form.noAck) payload.noAck = true
    if (form.allowRollup) payload.allowRollup = true
    if (form.allowDirect) payload.allowDirect = true
    if (form.mirrorDirect && form.mirrorEnabled) payload.mirrorDirect = true
    if (form.allowMsgCounter) payload.allowMsgCounter = true
    if (form.allowMsgSchedules) payload.allowMsgSchedules = true
    if (form.allowAtomicPublish) payload.allowAtomicPublish = true
    if (form.allowPerMessageTtl) payload.allowPerMessageTtl = true
    if (form.denyDelete) payload.denyDelete = true
    if (form.denyPurge) payload.denyPurge = true
    if (form.maxMsgs) payload.maxMsgs = parseInt(form.maxMsgs)
    if (form.maxBytes) payload.maxBytes = form.maxBytes
    if (form.maxAge) payload.maxAge = durationToGo(form.maxAge)
    if (form.maxMsgSize) payload.maxMsgSize = form.maxMsgSize
    if (form.maxMsgsPerSubject) payload.maxMsgsPerSubject = parseInt(form.maxMsgsPerSubject)
    if (form.maxConsumers) payload.maxConsumers = parseInt(form.maxConsumers)
    if (form.consumerInactiveThreshold) payload.consumerInactiveThreshold = durationToGo(form.consumerInactiveThreshold)
    if (form.maxAckPending) payload.maxAckPending = parseInt(form.maxAckPending)
    if (form.duplicates) payload.duplicates = durationToGo(form.duplicates)
    if (form.firstSeq) payload.firstSeq = parseInt(form.firstSeq)
    if (form.subjectDeleteMarkerTtl) payload.subjectDeleteMarkerTtl = durationToGo(form.subjectDeleteMarkerTtl)
    if (form.subjectTransformSrc || form.subjectTransformDest) {
      payload.subjectTransformSrc = form.subjectTransformSrc
      payload.subjectTransformDest = form.subjectTransformDest
    }
    if (form.metadata.length > 0) {
      payload.metadata = Object.fromEntries(form.metadata.filter(m => m.key).map(m => [m.key, m.value]))
    }
    if (form.republish && form.republishSrc && form.republishDest) {
      payload.republish = { src: form.republishSrc, dest: form.republishDest }
    }
    if (form.sources.length > 0) {
      payload.sources = form.sources.map(s => {
        const src: any = { name: s.name }
        if (s.filterSubject) src.filterSubject = s.filterSubject
        if (s.optStartSeq) src.optStartSeq = parseInt(s.optStartSeq)
        if (s.optStartTime) {
          const timePart = s.optStartTimeTime || '00:00:00'
          src.optStartTime = `${s.optStartTime}T${timePart}Z`
        }
        if (s.external) {
          src.external = {
            apiPrefix: s.apiPrefix,
            deliverPrefix: s.deliverPrefix,
          }
        }
        const transforms = s.subjectTransforms.filter(t => t.src || t.dest)
        if (transforms.length > 0) src.subjectTransforms = transforms
        return src
      })
    }
    if (form.tags.length > 0) payload.tags = form.tags
    if (form.persistMode !== 'default') payload.persistMode = form.persistMode

    if (isEditMode.value) {
      await streamsStore.updateStream(form.name, payload)
      showConfirm.value = false
      emit('updated', form.name)
    } else {
      await streamsStore.createStream(payload)
      showConfirm.value = false
      if (createAnother.value && !isMirrorMode.value && !isDuplicateMode.value) {
        resetForm()
      } else {
        emit('created', form.name)
      }
    }
  } catch (e: any) {
    serverError.value = e.message
    showConfirm.value = false
  } finally {
    saving.value = false
  }
}

// Toggle helper
function toggle(field: keyof typeof form) {
  ;(form as any)[field] = !(form as any)[field]
}
</script>

<template>
  <Teleport to="body">
    <!-- Config preview modal -->
    <div v-if="showConfirm" class="fixed inset-0 z-[60] flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="showConfirm = false"></div>
      <div class="relative w-full max-w-2xl max-h-[85vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">
              {{ isEditMode ? 'Confirm Stream Update' : isMirrorMode ? 'Confirm Mirror Creation' : 'Confirm Stream Creation' }}
            </h2>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">Please review the configuration that will be {{ isEditMode ? 'updated' : 'created' }}.</p>
          </div>
          <button class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400" @click="showConfirm = false">
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <div class="flex-1 overflow-y-auto p-4">
          <div class="bg-gray-900 rounded-lg overflow-hidden border border-gray-700">
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
          <button type="button" class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="showConfirm = false">Cancel</button>
          <button type="button" class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50" :disabled="saving" @click="submit">
            {{ saving ? (isEditMode ? 'Updating...' : 'Creating...') : (isEditMode ? 'Update Stream' : isMirrorMode ? 'Create Mirror' : 'Create Stream') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Main Modal -->
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="$emit('close')"></div>

      <div class="relative w-full max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-950 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-800 mx-4">
        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 shrink-0">
          <div v-if="isEditMode" class="flex items-center gap-1.5 text-sm text-gray-500 dark:text-gray-400">
            <span>Edit</span>
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5"/>
            </svg>
            <span class="font-semibold text-gray-900 dark:text-gray-100">{{ editStream!.config.name }}</span>
          </div>
          <h2 v-else-if="isMirrorMode" class="text-lg font-semibold text-gray-900 dark:text-gray-100">Create Mirror of {{ mirrorStream!.config.name }}</h2>
          <h2 v-else-if="isDuplicateMode" class="text-lg font-semibold text-gray-900 dark:text-gray-100">Duplicate {{ duplicateStream!.config.name }}</h2>
          <h2 v-else class="text-lg font-semibold text-gray-900 dark:text-gray-100">Create Stream</h2>
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
        <div class="flex-1 overflow-y-auto px-6 py-4 space-y-6">

          <!-- Basic Configuration (full width) -->
          <section>
            <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Basic Configuration</h3>
            <div class="grid grid-cols-2 gap-4">
              <!-- Name -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name <span class="text-red-500">*</span></label>
                <input
                  v-model="form.name"
                  type="text"
                  :readonly="isEditMode"
                  :class="[inputClass, fieldErrors.name ? 'border-red-400 dark:border-red-600' : '', isEditMode ? 'opacity-60 cursor-not-allowed' : '']"
                  placeholder="my-stream"
                  @input="delete (fieldErrors as any).name"
                />
                <p v-if="fieldErrors.name" class="mt-1 text-xs text-red-500">{{ fieldErrors.name }}</p>
              </div>
              <!-- Description -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
                <input v-model="form.description" type="text" :class="inputClass" placeholder="Optional description" />
              </div>
            </div>

            <!-- Mirror Stream toggle -->
            <div class="mt-3 flex items-center gap-2">
              <button type="button" @click="!isMirrorMode && toggle('mirrorEnabled')" :class="['relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent', form.mirrorEnabled ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-gray-700', isMirrorMode ? 'cursor-default' : 'cursor-pointer']" :disabled="isMirrorMode">
                <span :class="['pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform', form.mirrorEnabled ? 'translate-x-4' : 'translate-x-0']" />
              </button>
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Mirror Stream</span>
              <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Create a mirror that replicates all data from another stream">
                <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
              </span>
            </div>

            <!-- Mirror Settings -->
            <div v-if="form.mirrorEnabled" class="space-y-3 mt-3">
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Mirror Settings</h3>

              <!-- Source Stream (read-only) -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Source Stream</label>
                <input :value="form.mirrorName" type="text" :class="inputClass + ' bg-gray-100 dark:bg-gray-800 cursor-not-allowed'" readonly />
              </div>

              <!-- Start Sequence -->
              <div>
                <div class="flex items-center gap-1 mb-1">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Start Sequence</label>
                  <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Start mirroring from this sequence number">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                  </span>
                </div>
                <input v-model="form.mirrorStartSeq" type="text" :class="inputClass" placeholder="" />
              </div>

              <!-- Start Time -->
              <div>
                <div class="flex items-center gap-1 mb-1">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Start Time</label>
                  <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Start mirroring from this timestamp">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                  </span>
                </div>
                <div class="grid grid-cols-2 gap-3">
                  <input v-model="form.mirrorStartTime" type="date" :class="inputClass" />
                  <input v-model="form.mirrorStartTimeTime" type="time" step="1" :class="inputClass" placeholder="HH:MM:SS.mmm" />
                </div>
              </div>

              <!-- Filter Subject -->
              <div>
                <div class="flex items-center gap-1 mb-1">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Filter Subject</label>
                  <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Only mirror messages matching this subject">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                  </span>
                </div>
                <input v-model="form.mirrorFilterSubject" type="text" :class="inputClass" placeholder="orders.*" />
              </div>

              <!-- External Stream toggle -->
              <div>
                <div class="flex items-center gap-2 mb-2">
                  <button type="button" @click="form.mirrorExternal = !form.mirrorExternal" :class="['relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent cursor-pointer', form.mirrorExternal ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-gray-700']">
                    <span :class="['pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform', form.mirrorExternal ? 'translate-x-4' : 'translate-x-0']" />
                  </button>
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">External Stream</span>
                  <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Mirror from a stream in another account or domain">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                  </span>
                </div>
                <div v-if="form.mirrorExternal" class="space-y-3">
                  <div>
                    <div class="flex items-center gap-1 mb-1">
                      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">API Prefix</label>
                      <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Subject prefix for the remote API">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                      </span>
                    </div>
                    <input v-model="form.mirrorApiPrefix" type="text" :class="inputClass" placeholder="$JS.API" />
                  </div>
                  <div>
                    <div class="flex items-center gap-1 mb-1">
                      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Deliver Prefix</label>
                      <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Delivery subject for the push consumer">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                      </span>
                    </div>
                    <input v-model="form.mirrorDeliverPrefix" type="text" :class="inputClass" placeholder="$JS.INBOX" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Subjects -->
            <div v-if="!form.mirrorEnabled" class="mt-3">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subjects <span class="text-red-500">*</span></label>
              <input
                v-model="form.subjectInput"
                type="text"
                :class="[inputClass, fieldErrors.subjects ? 'border-red-400 dark:border-red-600' : '']"
                placeholder="Press space or enter to add (e.g. orders.*)"
                @keydown.enter.prevent="addSubject"
                @keydown.space.prevent="addSubject"
                @input="delete (fieldErrors as any).subjects"
              />
              <p v-if="fieldErrors.subjects" class="mt-1 text-xs text-red-500">{{ fieldErrors.subjects }}</p>
              <div v-if="form.subjects.length > 0" class="flex flex-wrap gap-1 mt-2">
                <span
                  v-for="s in form.subjects"
                  :key="s"
                  class="flex items-center gap-1 px-2 py-0.5 text-xs bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300 rounded-full font-mono"
                >
                  {{ s }}
                  <button type="button" class="hover:text-red-500" @click="removeSubject(s)">
                    <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/></svg>
                  </button>
                </span>
              </div>
            </div>

            <!-- Metadata -->
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

          <!-- Source Streams -->
          <section v-if="!form.mirrorEnabled">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">Source Streams
                <span class="inline-block ml-1 text-gray-400 dark:text-gray-600 cursor-help" title="Source messages from other streams into this stream">
                  <svg class="w-3.5 h-3.5 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                </span>
              </h3>
            </div>

            <!-- Source entries -->
            <div v-if="form.sources.length > 0" class="space-y-4">
              <div v-for="(source, si) in form.sources" :key="si" class="p-4 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 space-y-3">
                <!-- Header with Remove button -->
                <div class="flex items-center justify-end">
                  <button type="button" class="px-2 py-1 text-xs text-gray-500 hover:text-red-500 border border-gray-300 dark:border-gray-700 rounded hover:border-red-300 dark:hover:border-red-700" @click="removeSource(si)">Remove</button>
                </div>

                <!-- Stream Name -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Stream Name</label>
                  <input v-model="source.name" type="text" :class="[inputClass, fieldErrors[`source_${si}_name`] ? 'border-red-400 dark:border-red-600' : '']" placeholder="source-stream" @input="delete (fieldErrors as any)[`source_${si}_name`]" />
                  <p v-if="fieldErrors[`source_${si}_name`]" class="mt-1 text-xs text-red-500">{{ fieldErrors[`source_${si}_name`] }}</p>
                </div>

                <!-- Filter Subject + Start Sequence (side by side) -->
                <div class="grid grid-cols-2 gap-3">
                  <div>
                    <div class="flex items-center gap-1 mb-1">
                      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Filter Subject</label>
                      <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Only replicate messages matching this subject">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                      </span>
                    </div>
                    <input v-model="source.filterSubject" type="text" :class="inputClass" placeholder="source-stream.*" />
                  </div>
                  <div>
                    <div class="flex items-center gap-1 mb-1">
                      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Start Sequence</label>
                      <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Start sourcing from this sequence number">
                        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                      </span>
                    </div>
                    <input v-model="source.optStartSeq" type="text" :class="inputClass" placeholder="1" />
                  </div>
                </div>

                <!-- Start Time -->
                <div>
                  <div class="flex items-center gap-1 mb-1">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Start Time</label>
                    <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Start sourcing from this timestamp">
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                    </span>
                  </div>
                  <div class="grid grid-cols-2 gap-3">
                    <input v-model="source.optStartTime" type="date" :class="inputClass" />
                    <input v-model="source.optStartTimeTime" type="time" step="1" :class="inputClass" placeholder="HH:MM:SS.mmm" />
                  </div>
                </div>

                <!-- External Stream toggle -->
                <div>
                  <div class="flex items-center gap-2 mb-2">
                    <button type="button" @click="source.external = !source.external" :class="['relative inline-flex h-5 w-9 shrink-0 rounded-full border-2 border-transparent cursor-pointer', source.external ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-gray-700']">
                      <span :class="['pointer-events-none inline-block h-4 w-4 rounded-full bg-white shadow transform', source.external ? 'translate-x-4' : 'translate-x-0']" />
                    </button>
                    <span class="text-sm font-medium text-gray-700 dark:text-gray-300">External Stream</span>
                    <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Source from a stream in another account or domain">
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                    </span>
                  </div>
                  <div v-if="source.external" class="space-y-3 ml-0">
                    <div>
                      <div class="flex items-center gap-1 mb-1">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">API Prefix</label>
                        <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Subject prefix that imports the other account/domain API">
                          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                        </span>
                      </div>
                      <input v-model="source.apiPrefix" type="text" :class="inputClass" placeholder="$JS.API" />
                    </div>
                    <div>
                      <div class="flex items-center gap-1 mb-1">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Deliver Prefix</label>
                        <span class="text-gray-400 dark:text-gray-600 cursor-help" title="Delivery subject for the push consumer">
                          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4m0-4h.01"/></svg>
                        </span>
                      </div>
                      <input v-model="source.deliverPrefix" type="text" :class="inputClass" placeholder="$JS.INBOX" />
                    </div>
                  </div>
                </div>

                <!-- Subject Transforms -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subject Transforms</label>
                  <div v-if="source.subjectTransforms.length > 0" class="space-y-2 mb-2">
                    <div v-for="(t, ti) in source.subjectTransforms" :key="ti" class="flex gap-2 items-center">
                      <input v-model="t.src" type="text" :class="inputClass" placeholder="source.pattern.*" />
                      <input v-model="t.dest" type="text" :class="inputClass" placeholder="dest.pattern.{{1}}" />
                      <button type="button" class="text-gray-400 hover:text-red-500 shrink-0" @click="removeSourceTransform(si, ti)">
                        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                  <button type="button" class="text-xs text-emerald-600 dark:text-emerald-400 hover:underline" @click="addSourceTransform(si)">+ Add Transform</button>
                </div>
              </div>
            </div>

            <!-- Add Source button -->
            <button type="button" class="mt-2 flex items-center gap-1 text-xs text-emerald-600 dark:text-emerald-400 hover:underline" @click="addSource">
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
              </svg>
              Add Source
            </button>
          </section>

          <!-- Message Retention (left) + Infrastructure (right) -->
          <div class="grid grid-cols-2 gap-6">
            <!-- Message Retention -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Message Retention</h3>
              <div class="space-y-3">
                <!-- Retention Policy -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Retention Policy</label>
                  <select v-model="form.retention" :class="inputClass">
                    <option value="limits">Limits</option>
                    <option value="interest">Interest</option>
                    <option value="workqueue">Work Queue</option>
                  </select>
                </div>
                <!-- Discard Policy -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Discard Policy</label>
                  <div class="flex items-center gap-4">
                    <label class="flex items-center gap-2 cursor-pointer">
                      <input v-model="form.discard" type="radio" value="old" class="text-emerald-500 focus:ring-emerald-500" />
                      <span class="text-sm text-gray-700 dark:text-gray-300">Old</span>
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer">
                      <input v-model="form.discard" type="radio" value="new" class="text-emerald-500 focus:ring-emerald-500" />
                      <span class="text-sm text-gray-700 dark:text-gray-300">New</span>
                    </label>
                  </div>
                  <p v-if="fieldErrors.discard" class="mt-1 text-xs text-red-500">{{ fieldErrors.discard }}</p>
                </div>
                <!-- Discard New Per Subject -->
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-700 dark:text-gray-300">Discard New Per Subject</span>
                  <button type="button" class="relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none" :class="form.discardNewPerSubject ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-700'" @click="toggle('discardNewPerSubject')">
                    <span class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow" :class="form.discardNewPerSubject ? 'translate-x-4' : 'translate-x-0.5'"></span>
                  </button>
                </div>
              </div>
            </section>

            <!-- Infrastructure -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Infrastructure</h3>
              <div class="space-y-3">
                <!-- Desired Replicas -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Desired Replicas</label>
                  <div class="flex items-center gap-2">
                    <button type="button" class="w-7 h-7 flex items-center justify-center rounded border border-gray-300 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 text-sm" @click="form.replicas = Math.max(1, form.replicas - 1)">-</button>
                    <input v-model.number="form.replicas" type="number" min="1" max="5" class="w-16 px-2 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-emerald-500 text-center" />
                    <button type="button" class="w-7 h-7 flex items-center justify-center rounded border border-gray-300 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 text-sm" @click="form.replicas = Math.min(5, form.replicas + 1)">+</button>
                  </div>
                </div>
                <!-- Storage -->
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
                <!-- Compression -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Compression</label>
                  <select v-model="form.compression" :class="inputClass">
                    <option value="none">None</option>
                    <option value="s2">S2</option>
                  </select>
                </div>
                <!-- Persist Mode -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Persist Mode</label>
                  <select v-model="form.persistMode" :class="inputClass">
                    <option value="default">Default</option>
                    <option value="async">Async</option>
                  </select>
                  <p v-if="fieldErrors.persistMode" class="mt-1 text-xs text-red-500">{{ fieldErrors.persistMode }}</p>
                </div>
                <!-- Tags -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Tags</label>
                  <input
                    v-model="form.tagInput"
                    type="text"
                    :class="inputClass"
                    placeholder="Press space or enter to create a new tag"
                    @keydown.enter.prevent="addTag"
                    @keydown.space.prevent="addTag"
                  />
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

          <!-- Message Processing (left) + Stream Limits (right) -->
          <div class="grid grid-cols-2 gap-6">
            <!-- Message Processing -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Message Processing</h3>
              <div class="space-y-3">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">First Sequence</label>
                  <input v-model="form.firstSeq" type="number" :class="inputClass" placeholder="1" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Duplicate Window</label>
                  <input v-model="form.duplicates" type="text" :class="inputClass" placeholder="e.g. 2m, 1h" />
                  <p class="mt-1 text-xs text-gray-400 dark:text-gray-600">Duration: ns, us, ms, s, m, h, d</p>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subject Delete Marker TTL</label>
                  <input v-model="form.subjectDeleteMarkerTtl" type="text" :class="inputClass" placeholder="e.g. 5m" />
                  <p class="mt-1 text-xs text-gray-400 dark:text-gray-600">Duration: ns, us, ms, s, m, h, d</p>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subject Transform Source</label>
                  <input v-model="form.subjectTransformSrc" type="text" :class="inputClass" placeholder="input" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Subject Transform Destination</label>
                  <input v-model="form.subjectTransformDest" type="text" :class="inputClass" placeholder="e.g. output.$1" />
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-700 dark:text-gray-300">No Acknowledgement</span>
                  <button type="button" class="relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none" :class="form.noAck ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-700'" @click="toggle('noAck')">
                    <span class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow" :class="form.noAck ? 'translate-x-4' : 'translate-x-0.5'"></span>
                  </button>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-700 dark:text-gray-300">Allow Per-Message TTL</span>
                  <button type="button" class="relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none" :class="form.allowPerMessageTtl ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-700'" @click="toggle('allowPerMessageTtl')">
                    <span class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow" :class="form.allowPerMessageTtl ? 'translate-x-4' : 'translate-x-0.5'"></span>
                  </button>
                </div>
              </div>
            </section>

            <!-- Stream Limits -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Stream Limits</h3>
              <div class="space-y-3">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Message Age</label>
                  <input v-model="form.maxAge" type="text" :class="inputClass" placeholder="e.g. 24h, 7d" />
                  <p class="mt-1 text-xs text-gray-400 dark:text-gray-600">Duration: ns, us, ms, s, m, h, d</p>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Bytes Limit</label>
                  <input v-model="form.maxBytes" type="text" :class="inputClass" placeholder="e.g. 1GB, 100MB" />
                  <p class="mt-1 text-xs text-gray-400 dark:text-gray-600">Bytes (B), KB, MB, GB, TB or raw number</p>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Messages Limit</label>
                  <input v-model="form.maxMsgs" type="number" :class="inputClass" placeholder="Unlimited" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Messages Per Subject Limit</label>
                  <input v-model="form.maxMsgsPerSubject" type="number" :class="inputClass" placeholder="Unlimited" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Message Size Limit</label>
                  <input v-model="form.maxMsgSize" type="text" :class="inputClass" placeholder="e.g. 1MB, 64KB" />
                  <p class="mt-1 text-xs text-gray-400 dark:text-gray-600">Bytes (B), KB, MB, GB, TB or raw number</p>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Consumers Limit</label>
                  <input v-model="form.maxConsumers" type="number" :class="inputClass" placeholder="Unlimited" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Consumer Inactive Threshold</label>
                  <input v-model="form.consumerInactiveThreshold" type="text" :class="inputClass" placeholder="e.g. 5s" />
                  <p class="mt-1 text-xs text-gray-400 dark:text-gray-600">Duration: ns, us, ms, s, m, h, d</p>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Ack Pending</label>
                  <input v-model="form.maxAckPending" type="number" :class="inputClass" placeholder="Unlimited" />
                </div>
              </div>
            </section>
          </div>

          <!-- Advanced Features (left) + Republish Configuration (right) -->
          <div class="grid grid-cols-2 gap-6">
            <!-- Advanced Features -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Advanced Features</h3>
              <div class="space-y-2.5">
                <div v-for="item in ([
                  ['allowRollup', 'Allow Rollup Headers'],
                  ['allowDirect', 'Allow Direct Access'],
                  ['allowMsgCounter', 'Allow Message Counter'],
                  ['allowMsgSchedules', 'Allow Message Schedules'],
                  ['allowAtomicPublish', 'Allow Atomic Publish'],
                  ['denyDelete', 'Deny Delete'],
                  ['denyPurge', 'Deny Purge'],
                ])" :key="item[0]" class="flex items-center justify-between">
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ item[1] }}</span>
                  <button type="button" class="relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none" :class="(form as any)[item[0]] ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-700'" @click="(form as any)[item[0]] = !(form as any)[item[0]]">
                    <span class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow" :class="(form as any)[item[0]] ? 'translate-x-4' : 'translate-x-0.5'"></span>
                  </button>
                </div>
                <p v-if="fieldErrors.denyPurge" class="text-xs text-red-500">{{ fieldErrors.denyPurge }}</p>
                <div class="flex items-center justify-between" :class="!form.mirrorEnabled ? 'opacity-50' : ''">
                  <div class="flex items-center gap-1">
                    <span class="text-sm text-gray-700 dark:text-gray-300">Mirror Direct Access</span>
                    <span v-if="!form.mirrorEnabled" class="text-xs text-gray-400 dark:text-gray-600">(requires mirror)</span>
                  </div>
                  <button type="button" :disabled="!form.mirrorEnabled" :class="['relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none', form.mirrorEnabled ? (form.mirrorDirect ? 'bg-emerald-600 cursor-pointer' : 'bg-gray-300 dark:bg-gray-700 cursor-pointer') : 'bg-gray-300 dark:bg-gray-700 cursor-not-allowed']" @click="form.mirrorEnabled && toggle('mirrorDirect')">
                    <span class="inline-block h-3.5 w-3.5 rounded-full bg-white shadow" :class="form.mirrorDirect ? 'translate-x-4' : 'translate-x-0.5'"></span>
                  </button>
                </div>
              </div>
            </section>

            <!-- Republish Configuration -->
            <section>
              <h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-3">Republish Configuration</h3>
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-gray-700 dark:text-gray-300">Enable Republish</span>
                  <button type="button" class="relative inline-flex h-5 w-9 items-center rounded-full border-2 border-transparent focus:outline-none" :class="form.republish ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-700'" @click="toggle('republish')">
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
              </div>
            </section>
          </div>

        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-800 shrink-0 flex items-center gap-3">
          <!-- Show Config / Show Config Diff button -->
          <button type="button" class="flex items-center gap-1.5 px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="handleShowConfig">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5"/></svg>
            {{ isEditMode ? 'Show Config Diff' : 'Show Config' }}
          </button>
          <!-- Create another (only in create mode) -->
          <label v-if="!isEditMode && !isMirrorMode && !isDuplicateMode" class="flex items-center gap-2 cursor-pointer select-none">
            <input v-model="createAnother" type="checkbox" class="rounded border-gray-300 dark:border-gray-700 text-emerald-600 focus:ring-emerald-500" />
            <span class="text-sm text-gray-700 dark:text-gray-300">Create another</span>
          </label>
          <div class="flex-1"></div>
          <button type="button" class="px-4 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800" @click="$emit('close')">Cancel</button>
          <button type="button" class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 rounded-md hover:bg-emerald-700 disabled:opacity-50" :disabled="saving" @click="submit">
            {{ saving ? (isEditMode ? 'Updating...' : 'Creating...') : (isEditMode ? 'Update Stream' : isMirrorMode ? 'Create Mirror' : 'Create Stream') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
