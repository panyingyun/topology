import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN'
import enUS from './en-US'

export type SupportedLocale = 'zh-CN' | 'en-US'

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

// Get saved locale from localStorage or default to 'zh-CN'
const getSavedLocale = (): SupportedLocale => {
  const saved = localStorage.getItem('topology-locale')
  if (saved && (saved === 'zh-CN' || saved === 'en-US')) {
    return saved as SupportedLocale
  }
  // Try to detect from browser
  const browserLang = navigator.language || (navigator as any).userLanguage
  if (browserLang.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en-US'
}

const i18n = createI18n({
  legacy: false,
  locale: getSavedLocale(),
  fallbackLocale: 'en-US',
  messages,
})

export default i18n

export const setLocale = (locale: SupportedLocale) => {
  i18n.global.locale.value = locale
  localStorage.setItem('topology-locale', locale)
}

export const getLocale = (): SupportedLocale => {
  return i18n.global.locale.value as SupportedLocale
}
