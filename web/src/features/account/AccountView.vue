<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useConnectionsStore } from '../../stores/connections'
import { serverApi } from '../../lib/api'

const connStore = useConnectionsStore()
const info = ref<any>(null)
const error = ref('')

async function fetchData() {
  if (!connStore.status.connected) return
  error.value = ''
  try {
    info.value = await serverApi.accountInfo()
  } catch (e: any) {
    error.value = e.message
  }
}

watch(() => connStore.status.connectionId, () => { fetchData() })
onMounted(() => { if (connStore.status.connected) fetchData() })

// ── Formatting helpers ────────────────────────────────────────────────────────

const UNLIMITED_THRESHOLD = 2 ** 60   // anything >= this is treated as unlimited
const NEG_ONE = -1

function isUnlimited(v: number | undefined | null): boolean {
  if (v === undefined || v === null) return true
  return v === NEG_ONE || v >= UNLIMITED_THRESHOLD
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(Math.abs(bytes)) / Math.log(k))
  const idx = Math.min(i, sizes.length - 1)
  const val = bytes / Math.pow(k, idx)
  return val % 1 === 0 ? `${val} ${sizes[idx]}` : `${val.toFixed(1)} ${sizes[idx]}`
}

function formatNumber(n: number): string {
  return n.toLocaleString()
}

function pct(current: number, limit: number): number {
  if (isUnlimited(limit) || limit === 0) return 0
  return Math.min(100, Math.round((current / limit) * 100))
}

function limitLabel(limit: number, formatter: (v: number) => string): string {
  return isUnlimited(limit) ? 'unlimited' : formatter(limit)
}
</script>

<template>
  <div
    v-if="!connStore.status.connected"
    class="flex items-center justify-center h-full text-gray-500 dark:text-gray-400 text-sm"
  >
    Connect to a NATS server to view account information.
  </div>

  <div v-else class="p-6 space-y-6 max-w-6xl">
    <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">Account</h1>

    <div
      v-if="error"
      class="p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/50 dark:text-red-400 rounded-md border border-red-200 dark:border-red-800"
    >
      {{ error }}
    </div>

    <template v-if="info">

      <!-- ── Account Information ──────────────────────────────────────────── -->
      <section>
        <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">
          Account Information
        </h2>
        <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
          <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">Server Details</h3>
          <div class="space-y-2">

            <div class="flex items-baseline gap-1">
              <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">SERVER NAME</span>
              <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
              <span class="font-mono text-xs text-gray-900 dark:text-gray-100 shrink-0">{{ info.server_name || '—' }}</span>
            </div>

            <div class="flex items-baseline gap-1">
              <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">SERVER VERSION</span>
              <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
              <span class="font-mono text-xs text-gray-900 dark:text-gray-100 shrink-0">{{ info.server_version || '—' }}</span>
            </div>

            <div class="flex items-baseline gap-1">
              <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">MAX MESSAGE PAYLOAD</span>
              <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
              <span class="font-mono text-xs text-gray-900 dark:text-gray-100 shrink-0">{{ formatBytes(info.max_payload ?? 0) }}</span>
            </div>

            <div class="flex items-baseline gap-1">
              <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">JETSTREAM ENABLED</span>
              <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
              <span
                class="font-mono text-xs shrink-0"
                :class="info.jetstream ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-500 dark:text-gray-400'"
              >{{ info.jetstream ? 'true' : 'false' }}</span>
            </div>

            <div class="flex items-baseline gap-1">
              <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">TOTAL API CALLS</span>
              <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
              <span class="font-mono text-xs text-gray-900 dark:text-gray-100 shrink-0">{{ formatNumber(info.api_total ?? 0) }}</span>
            </div>

            <div class="flex items-baseline gap-1">
              <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">ERROR API CALLS</span>
              <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
              <span
                class="font-mono text-xs shrink-0"
                :class="(info.api_errors ?? 0) > 0 ? 'text-red-500 dark:text-red-400' : 'text-gray-900 dark:text-gray-100'"
              >{{ formatNumber(info.api_errors ?? 0) }}</span>
            </div>

          </div>
        </div>
      </section>

      <!-- ── JetStream section ───────────────────────────────────────────── -->
      <section v-if="info.jetstream">
        <h2 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-3">
          JetStream
        </h2>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

          <!-- Resource Usage -->
          <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">Resource Usage</h3>
            <div class="space-y-4">

              <!-- STREAMS -->
              <div>
                <div class="flex items-baseline justify-between mb-1">
                  <span class="text-sm text-gray-500 dark:text-gray-400">STREAMS</span>
                  <span class="font-mono text-xs text-gray-900 dark:text-gray-100">
                    {{ formatNumber(info.streams ?? 0) }}
                    <span class="text-gray-500 dark:text-gray-400"> / </span>
                    {{ limitLabel(info.limits?.max_streams, formatNumber) }}
                    <span
                      v-if="!isUnlimited(info.limits?.max_streams)"
                      class="text-gray-500 dark:text-gray-400"
                    > ({{ pct(info.streams ?? 0, info.limits?.max_streams) }}%)</span>
                  </span>
                </div>
                <div v-if="!isUnlimited(info.limits?.max_streams)" class="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-1.5">
                  <div
                    class="bg-emerald-500 h-1.5 rounded-full"
                    :style="{ width: pct(info.streams ?? 0, info.limits?.max_streams) + '%' }"
                  ></div>
                </div>
              </div>

              <!-- CONSUMERS -->
              <div>
                <div class="flex items-baseline justify-between mb-1">
                  <span class="text-sm text-gray-500 dark:text-gray-400">CONSUMERS</span>
                  <span class="font-mono text-xs text-gray-900 dark:text-gray-100">
                    {{ formatNumber(info.consumers ?? 0) }}
                    <span class="text-gray-500 dark:text-gray-400"> / </span>
                    {{ limitLabel(info.limits?.max_consumers, formatNumber) }}
                    <span
                      v-if="!isUnlimited(info.limits?.max_consumers)"
                      class="text-gray-500 dark:text-gray-400"
                    > ({{ pct(info.consumers ?? 0, info.limits?.max_consumers) }}%)</span>
                  </span>
                </div>
                <div v-if="!isUnlimited(info.limits?.max_consumers)" class="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-1.5">
                  <div
                    class="bg-emerald-500 h-1.5 rounded-full"
                    :style="{ width: pct(info.consumers ?? 0, info.limits?.max_consumers) + '%' }"
                  ></div>
                </div>
              </div>

              <!-- FILE STORAGE -->
              <div>
                <div class="flex items-baseline justify-between mb-1">
                  <span class="text-sm text-gray-500 dark:text-gray-400">FILE STORAGE</span>
                  <span class="font-mono text-xs text-gray-900 dark:text-gray-100">
                    {{ formatBytes(info.storage ?? 0) }}
                    <span class="text-gray-500 dark:text-gray-400"> / </span>
                    {{ limitLabel(info.limits?.max_storage, formatBytes) }}
                    <span
                      v-if="!isUnlimited(info.limits?.max_storage)"
                      class="text-gray-500 dark:text-gray-400"
                    > ({{ pct(info.storage ?? 0, info.limits?.max_storage) }}%)</span>
                  </span>
                </div>
                <div v-if="!isUnlimited(info.limits?.max_storage)" class="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-1.5">
                  <div
                    class="bg-emerald-500 h-1.5 rounded-full"
                    :style="{ width: pct(info.storage ?? 0, info.limits?.max_storage) + '%' }"
                  ></div>
                </div>
              </div>

              <!-- MEMORY STORAGE -->
              <div>
                <div class="flex items-baseline justify-between mb-1">
                  <span class="text-sm text-gray-500 dark:text-gray-400">MEMORY STORAGE</span>
                  <span class="font-mono text-xs text-gray-900 dark:text-gray-100">
                    {{ formatBytes(info.memory ?? 0) }}
                    <span class="text-gray-500 dark:text-gray-400"> / </span>
                    {{ limitLabel(info.limits?.max_memory, formatBytes) }}
                    <span
                      v-if="!isUnlimited(info.limits?.max_memory)"
                      class="text-gray-500 dark:text-gray-400"
                    > ({{ pct(info.memory ?? 0, info.limits?.max_memory) }}%)</span>
                  </span>
                </div>
                <div v-if="!isUnlimited(info.limits?.max_memory)" class="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-1.5">
                  <div
                    class="bg-emerald-500 h-1.5 rounded-full"
                    :style="{ width: pct(info.memory ?? 0, info.limits?.max_memory) + '%' }"
                  ></div>
                </div>
              </div>

            </div>
          </div>

          <!-- Configuration Requirements & Limits -->
          <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4">
            <h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-4">Configuration Requirements &amp; Limits</h3>
            <div class="space-y-2">

              <div class="flex items-baseline gap-1">
                <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">FILE STORAGE PER STREAM</span>
                <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
                <span class="font-mono text-xs text-gray-900 dark:text-gray-100 shrink-0">
                  {{ limitLabel(info.limits?.store_max_stream_bytes, formatBytes) }}
                </span>
              </div>

              <div class="flex items-baseline gap-1">
                <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">MEMORY STORAGE PER STREAM</span>
                <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
                <span class="font-mono text-xs text-gray-900 dark:text-gray-100 shrink-0">
                  {{ limitLabel(info.limits?.memory_max_stream_bytes, formatBytes) }}
                </span>
              </div>

              <div class="flex items-baseline gap-1">
                <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">STREAM REQUIRES MAX BYTES SET</span>
                <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
                <span
                  class="font-mono text-xs shrink-0"
                  :class="info.limits?.max_bytes_required ? 'text-amber-600 dark:text-amber-400' : 'text-gray-900 dark:text-gray-100'"
                >{{ info.limits?.max_bytes_required ? 'true' : 'false' }}</span>
              </div>

              <div class="flex items-baseline gap-1">
                <span class="text-sm text-gray-500 dark:text-gray-400 shrink-0">CONSUMER MAXIMUM ACK PENDING</span>
                <span class="flex-1 border-b border-dotted border-gray-300 dark:border-gray-700 mb-0.5 mx-1"></span>
                <span class="font-mono text-xs text-gray-900 dark:text-gray-100 shrink-0">
                  {{ isUnlimited(info.limits?.max_ack_pending) ? 'unlimited' : formatNumber(info.limits?.max_ack_pending) }}
                </span>
              </div>

            </div>
          </div>

        </div>
      </section>

    </template>

    <!-- Loading skeleton while waiting for first data -->
    <template v-else-if="!error">
      <div class="space-y-3 animate-pulse">
        <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-1/4"></div>
        <div class="h-32 bg-gray-100 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800"></div>
        <div class="h-4 bg-gray-200 dark:bg-gray-800 rounded w-1/4 mt-4"></div>
        <div class="grid grid-cols-2 gap-6">
          <div class="h-40 bg-gray-100 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800"></div>
          <div class="h-40 bg-gray-100 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800"></div>
        </div>
      </div>
    </template>

  </div>
</template>
