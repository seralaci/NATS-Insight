<script setup lang="ts">
import { computed, ref } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import type { MetricPoint } from '../../composables/useMetrics'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{
  title: string
  points: MetricPoint[]
  dataKeys: [keyof MetricPoint, keyof MetricPoint] | [keyof MetricPoint]
  labels: [string, string] | [string]
  colors: [string, string] | [string]
  unit?: string
}>()

const isDark = ref(document.documentElement.classList.contains('dark'))

const observer = new MutationObserver(() => {
  isDark.value = document.documentElement.classList.contains('dark')
})
observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })

import { onUnmounted } from 'vue'
onUnmounted(() => observer.disconnect())

function formatTime(ts: number): string {
  const d = new Date(ts)
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  const ss = String(d.getSeconds()).padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}

const chartData = computed(() => {
  const pts = props.points
  const labels = pts.map((p) => formatTime(p.timestamp))

  const datasets = props.dataKeys.map((key, i) => ({
    label: props.labels[i],
    data: pts.map((p) => p[key] as number),
    borderColor: props.colors[i],
    backgroundColor: props.colors[i] + '18',
    borderWidth: 1.5,
    tension: 0.3,
    pointRadius: 0,
    pointHoverRadius: 3,
    fill: props.dataKeys.length === 1,
  }))

  return { labels, datasets }
})

const chartOptions = computed(() => {
  const gridColor = isDark.value ? 'rgba(75,85,99,0.3)' : 'rgba(209,213,219,0.5)'
  const tickColor = isDark.value ? '#9ca3af' : '#6b7280'

  return {
    responsive: true,
    maintainAspectRatio: false,
    animation: false as const,
    interaction: {
      mode: 'index' as const,
      intersect: false,
    },
    plugins: {
      legend: {
        position: 'top' as const,
        labels: {
          color: tickColor,
          font: { size: 11 },
          boxWidth: 12,
          padding: 8,
        },
      },
      tooltip: {
        callbacks: {
          label: (ctx: any) => {
            const val = ctx.parsed.y
            if (props.unit) return `${ctx.dataset.label}: ${val} ${props.unit}`
            return `${ctx.dataset.label}: ${val}`
          },
        },
      },
    },
    scales: {
      x: {
        ticks: {
          color: tickColor,
          font: { size: 10 },
          maxTicksLimit: 6,
          maxRotation: 0,
        },
        grid: {
          color: gridColor,
        },
      },
      y: {
        beginAtZero: true,
        ticks: {
          color: tickColor,
          font: { size: 10 },
          maxTicksLimit: 5,
        },
        grid: {
          color: gridColor,
        },
      },
    },
  }
})
</script>

<template>
  <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4 flex flex-col gap-2">
    <div class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{{ title }}</div>
    <div style="height: 200px; position: relative;">
      <Line v-if="points.length > 0" :data="chartData" :options="chartOptions" />
      <div v-else class="flex items-center justify-center h-full">
        <span class="text-xs text-gray-400 dark:text-gray-600">Waiting for data...</span>
      </div>
    </div>
  </div>
</template>
