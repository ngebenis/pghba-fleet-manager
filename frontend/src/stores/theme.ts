import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDark = ref(localStorage.getItem('theme') !== 'light')

  function applyTheme(dark: boolean) {
    const root = document.documentElement
    if (dark) {
      root.style.setProperty('--c-bg',       '#0f1117')
      root.style.setProperty('--c-surface',  '#1a1d27')
      root.style.setProperty('--c-surface2', '#21253a')
      root.style.setProperty('--c-border',   '#2a2d3e')
      root.style.setProperty('--c-border2',  '#343749')
      root.style.setProperty('--c-text',     '#e2e8f0')
      root.style.setProperty('--c-muted',    '#7c85a0')
    } else {
      root.style.setProperty('--c-bg',       '#f0f2f7')
      root.style.setProperty('--c-surface',  '#ffffff')
      root.style.setProperty('--c-surface2', '#f5f7fb')
      root.style.setProperty('--c-border',   '#dde1ed')
      root.style.setProperty('--c-border2',  '#c8cedd')
      root.style.setProperty('--c-text',     '#1a1d2e')
      root.style.setProperty('--c-muted',    '#6b7280')
    }
  }

  function toggle() {
    isDark.value = !isDark.value
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
    applyTheme(isDark.value)
  }

  // Apply on init
  applyTheme(isDark.value)

  return { isDark, toggle, applyTheme }
})
