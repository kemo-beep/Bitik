export type Money = {
  amount: number | string
  currency?: string
}

let DEFAULT_LOCALE = "en-US"
const DEFAULT_CURRENCY = "MMK"

export function formatMoney(value: Money | number | string, opts?: {
  locale?: string
  currency?: string
  compact?: boolean
}): string {
  const locale = opts?.locale ?? DEFAULT_LOCALE
  let amount: number
  let currency: string

  if (typeof value === "object" && value !== null) {
    amount = typeof value.amount === "string" ? Number(value.amount) : value.amount
    currency = value.currency ?? opts?.currency ?? DEFAULT_CURRENCY
  } else {
    amount = typeof value === "string" ? Number(value) : value
    currency = opts?.currency ?? DEFAULT_CURRENCY
  }

  if (!Number.isFinite(amount)) return "—"

  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    notation: opts?.compact ? "compact" : "standard",
    maximumFractionDigits: opts?.compact ? 1 : 2,
  }).format(amount)
}

export function formatMoneyRange(min: Money | number, max: Money | number, opts?: { locale?: string; currency?: string }): string {
  const a = formatMoney(min, opts)
  const b = formatMoney(max, opts)
  return a === b ? a : `${a} – ${b}`
}

export function formatNumber(value: number, opts?: Intl.NumberFormatOptions & { locale?: string }): string {
  if (!Number.isFinite(value)) return "—"
  const { locale, ...rest } = opts ?? {}
  return new Intl.NumberFormat(locale ?? DEFAULT_LOCALE, rest).format(value)
}

export function formatDate(value: string | Date | number, opts?: Intl.DateTimeFormatOptions & { locale?: string }): string {
  const d = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(d.getTime())) return "—"
  const { locale, ...rest } = opts ?? {}
  return new Intl.DateTimeFormat(locale ?? DEFAULT_LOCALE, {
    year: "numeric",
    month: "short",
    day: "2-digit",
    ...rest,
  }).format(d)
}

export function formatDateTime(value: string | Date | number, opts?: { locale?: string }): string {
  return formatDate(value, {
    ...opts,
    hour: "2-digit",
    minute: "2-digit",
  })
}

const RELATIVE_UNITS: [Intl.RelativeTimeFormatUnit, number][] = [
  ["year", 60 * 60 * 24 * 365],
  ["month", 60 * 60 * 24 * 30],
  ["week", 60 * 60 * 24 * 7],
  ["day", 60 * 60 * 24],
  ["hour", 60 * 60],
  ["minute", 60],
  ["second", 1],
]

export function formatRelative(value: string | Date | number, opts?: { locale?: string; now?: Date }): string {
  const d = value instanceof Date ? value : new Date(value)
  if (Number.isNaN(d.getTime())) return "—"
  const now = opts?.now ?? new Date()
  const diff = (d.getTime() - now.getTime()) / 1000
  const fmt = new Intl.RelativeTimeFormat(opts?.locale ?? DEFAULT_LOCALE, { numeric: "auto" })
  for (const [unit, seconds] of RELATIVE_UNITS) {
    if (Math.abs(diff) >= seconds || unit === "second") {
      return fmt.format(Math.round(diff / seconds), unit)
    }
  }
  return fmt.format(0, "second")
}

export function formatPercent(value: number, opts?: { locale?: string; fraction?: number }): string {
  if (!Number.isFinite(value)) return "—"
  return new Intl.NumberFormat(opts?.locale ?? DEFAULT_LOCALE, {
    style: "percent",
    maximumFractionDigits: opts?.fraction ?? 1,
  }).format(value)
}

export function setDefaultFormatLocale(locale: string) {
  DEFAULT_LOCALE = locale
}

export function getDefaultFormatLocale() {
  return DEFAULT_LOCALE
}

export function truncate(value: string, max: number): string {
  if (value.length <= max) return value
  return value.slice(0, Math.max(0, max - 1)).trimEnd() + "…"
}

export function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("")
}
