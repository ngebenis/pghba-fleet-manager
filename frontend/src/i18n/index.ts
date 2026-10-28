import { reactive, computed, watchEffect } from 'vue'
import en from './en'
import id from './id'
import zh from './zh'
import es from './es'
import ar from './ar'
import fr from './fr'
import pt from './pt'
import bn from './bn'
import ru from './ru'
import hi from './hi'
import th from './th'
import km from './km'
import tl from './tl'
import vi from './vi'
import ja from './ja'
import ko from './ko'
import pl from './pl'
import it from './it'
import tr from './tr'
import ta from './ta'
import fa from './fa'
import uk from './uk'
import de from './de'
import nl from './nl'

export type Locale =
  | 'en' | 'id' | 'zh' | 'es' | 'ar' | 'fr' | 'pt' | 'bn' | 'ru' | 'hi'
  | 'th' | 'km' | 'tl' | 'vi' | 'ja' | 'ko' | 'pl' | 'it' | 'tr' | 'ta'
  | 'fa' | 'uk' | 'de' | 'nl'
type Messages = typeof en

const messages: Record<Locale, Messages> = {
  en, id, zh, es, ar, fr, pt, bn, ru, hi, th, km, tl, vi, ja, ko, pl, it, tr, ta, fa, uk, de, nl,
}

export interface LocaleInfo { code: Locale; name: string; flag: string; rtl?: boolean }

export const AVAILABLE_LOCALES: LocaleInfo[] = [
  { code: 'en', name: 'English', flag: '🇬🇧' },
  { code: 'id', name: 'Bahasa Indonesia', flag: '🇮🇩' },
  { code: 'zh', name: '中文', flag: '🇨🇳' },
  { code: 'es', name: 'Español', flag: '🇪🇸' },
  { code: 'ar', name: 'العربية', flag: '🇸🇦', rtl: true },
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
  { code: 'pt', name: 'Português', flag: '🇵🇹' },
  { code: 'bn', name: 'বাংলা', flag: '🇧🇩' },
  { code: 'ru', name: 'Русский', flag: '🇷🇺' },
  { code: 'hi', name: 'हिन्दी', flag: '🇮🇳' },
  { code: 'th', name: 'ไทย', flag: '🇹🇭' },
  { code: 'km', name: 'ខ្មែរ', flag: '🇰🇭' },
  { code: 'tl', name: 'Tagalog', flag: '🇵🇭' },
  { code: 'vi', name: 'Tiếng Việt', flag: '🇻🇳' },
  { code: 'ja', name: '日本語', flag: '🇯🇵' },
  { code: 'ko', name: '한국어', flag: '🇰🇷' },
  { code: 'pl', name: 'Polski', flag: '🇵🇱' },
  { code: 'it', name: 'Italiano', flag: '🇮🇹' },
  { code: 'tr', name: 'Türkçe', flag: '🇹🇷' },
  { code: 'ta', name: 'தமிழ்', flag: '🇮🇳' },
  { code: 'fa', name: 'فارسی', flag: '🇮🇷', rtl: true },
  { code: 'uk', name: 'Українська', flag: '🇺🇦' },
  { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
  { code: 'nl', name: 'Nederlands', flag: '🇳🇱' },
]

const RTL_LOCALES = new Set<Locale>(AVAILABLE_LOCALES.filter(l => l.rtl).map(l => l.code))

const validLocales = new Set(AVAILABLE_LOCALES.map(l => l.code))
const storedLocale = localStorage.getItem('locale') as Locale | null

const state = reactive({
  locale: (storedLocale && validLocales.has(storedLocale) ? storedLocale : 'en') as Locale,
})

watchEffect(() => {
  document.documentElement.dir = RTL_LOCALES.has(state.locale) ? 'rtl' : 'ltr'
  document.documentElement.lang = state.locale
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

  return {
    t,
    locale: computed(() => state.locale),
    isRTL: computed(() => RTL_LOCALES.has(state.locale)),
    setLocale,
    $t,
  }
}
