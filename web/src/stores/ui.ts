import { defineStore } from 'pinia'
import { ref } from 'vue'

const SIDEBAR_MIN = 160
const SIDEBAR_MAX = 480

export const useUiStore = defineStore('ui', () => {
  const sidebarCollapsed = ref(false)
  const theme = ref<'light' | 'dark' | 'system'>('system')
  const sidebarWidth = ref(208)

  function setSidebarWidth(w: number) {
    sidebarWidth.value = Math.min(SIDEBAR_MAX, Math.max(SIDEBAR_MIN, w))
    localStorage.setItem('sidebarWidth', String(sidebarWidth.value))
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setTheme(t: 'light' | 'dark' | 'system') {
    theme.value = t
    localStorage.setItem('theme', t)
    applyTheme()
  }

  function applyTheme() {
    const root = document.documentElement
    if (theme.value === 'dark' || (theme.value === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
  }

  function init() {
    const saved = localStorage.getItem('theme') as 'light' | 'dark' | 'system' | null
    if (saved) theme.value = saved
    applyTheme()
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme)
    const savedWidth = localStorage.getItem('sidebarWidth')
    if (savedWidth) sidebarWidth.value = Math.min(SIDEBAR_MAX, Math.max(SIDEBAR_MIN, Number(savedWidth)))
  }

  return { sidebarCollapsed, theme, sidebarWidth, toggleSidebar, setTheme, setSidebarWidth, init }
})
