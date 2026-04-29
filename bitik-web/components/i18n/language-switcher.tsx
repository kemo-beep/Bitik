"use client"

import { useI18n } from "@/lib/i18n/context"

export function LanguageSwitcher() {
  const { locale, setLocale, t } = useI18n()
  return (
    <label className="inline-flex items-center gap-2 text-xs text-muted-foreground">
      <span>{t("common.language")}</span>
      <select
        aria-label={t("common.language")}
        className="h-8 rounded border bg-background px-2 text-xs text-foreground"
        value={locale}
        onChange={(e) => {
          const v = e.target.value
          if (v === "ar") setLocale("ar")
          else if (v === "fr") setLocale("fr")
          else setLocale("en")
        }}
      >
        <option value="en">EN</option>
        <option value="fr">FR</option>
        <option value="ar">AR</option>
      </select>
    </label>
  )
}

