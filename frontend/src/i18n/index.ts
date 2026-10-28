import { reactive, computed } from 'vue'
import en from './en'
import id from './id'

type Locale = 'en' | 'id'
type Messages = typeof en

const messages: Record<Locale, Messages> = { en, id }

const state = reactive({
  locale: (localStorage.getItem('locale') as Locale) || 'en',
})

export function useI18n() {
  const t = computed(() => messages[state.locale])

  function setLocale(locale: Locale) {
    state.locale = locale
    localStorage.setItem('locale', locale)
  }

  function $t(key: string): string {
    const keys = key.split('.')
    let result: any = messages[state.locale]
    for (const k of keys) {
      result = result?.[k]
      if (result === undefined) return key
    }
    return result || key
  }

  return { t, locale: computed(() => state.locale), setLocale, $t }
}
