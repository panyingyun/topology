import { ref, watch, onMounted } from 'vue'

const STORAGE_KEY = 'app-theme'
export type Theme = 'light' | 'dark'

const theme = ref<Theme>((localStorage.getItem(STORAGE_KEY) as Theme) || 'dark')

function applyTheme(value: Theme) {
  const root = document.documentElement
  root.setAttribute('data-theme', value)
  root.classList.remove('theme-light', 'theme-dark')
  root.classList.add(`theme-${value}`)
  try {
    if (typeof window !== 'undefined' && (window as any).runtime?.WindowSetDarkTheme) {
      if (value === 'dark') (window as any).runtime.WindowSetDarkTheme()
      else (window as any).runtime.WindowSetLightTheme()
    }
  } catch {
    // ignore Wails theme API
  }
}

export function useTheme() {
  const setTheme = (value: Theme) => {
    theme.value = value
    localStorage.setItem(STORAGE_KEY, value)
    applyTheme(value)
  }

  const toggleTheme = () => {
    setTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  watch(theme, applyTheme, { immediate: false })

  onMounted(() => {
    applyTheme(theme.value)
  })

  return { theme, setTheme, toggleTheme }
}
