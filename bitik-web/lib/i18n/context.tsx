"use client"

import * as React from "react"
import { localeDirection, messages, type Locale } from "@/lib/i18n/messages"
import { setDefaultFormatLocale } from "@/lib/format"

const STORAGE_KEY = "bitik.locale.v1"

type I18nContextValue = {
  locale: Locale
  direction: "ltr" | "rtl"
  setLocale: (next: Locale) => void
  t: (key: string) => string
}

const I18nContext = React.createContext<I18nContextValue>({
  locale: "en",
  direction: "ltr",
  setLocale: () => {},
  t: (key) => key,
})

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = React.useState<Locale>("en")

  React.useEffect(() => {
    const stored = typeof window !== "undefined" ? window.localStorage.getItem(STORAGE_KEY) : null
    if (stored === "en" || stored === "ar") setLocaleState(stored)
  }, [])

  React.useEffect(() => {
    if (typeof document !== "undefined") {
      document.documentElement.lang = locale
      document.documentElement.dir = localeDirection[locale]
    }
    if (typeof window !== "undefined") {
      window.localStorage.setItem(STORAGE_KEY, locale)
    }
    setDefaultFormatLocale(locale === "ar" ? "ar" : "en-US")
  }, [locale])

  const setLocale = React.useCallback((next: Locale) => setLocaleState(next), [])
  const t = React.useCallback((key: string) => messages[locale][key] ?? messages.en[key] ?? key, [locale])

  const value = React.useMemo<I18nContextValue>(
    () => ({ locale, direction: localeDirection[locale], setLocale, t }),
    [locale, setLocale, t]
  )

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}

export function useI18n() {
  return React.useContext(I18nContext)
}

