import type { MetadataRoute } from "next"
import { env } from "@/lib/env"

type Envelope<T> = { success?: boolean; data?: T }

function asRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null
}

function asArray(value: unknown): unknown[] | null {
  return Array.isArray(value) ? value : null
}

function asString(value: unknown): string | null {
  return typeof value === "string" ? value : null
}

function asNumber(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null
}

async function fetchEnvelope<T>(url: string): Promise<T | null> {
  try {
    const res = await fetch(url, { method: "GET" })
    if (!res.ok) return null
    const json = (await res.json()) as Envelope<T>
    return (json && typeof json === "object" ? json.data : null) ?? null
  } catch {
    return null
  }
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base = env.oauthRedirectBaseUrl
  const now = new Date()

  const out: MetadataRoute.Sitemap = [
    { url: `${base}/`, lastModified: now, changeFrequency: "daily", priority: 1 },
    { url: `${base}/products`, lastModified: now, changeFrequency: "daily", priority: 0.8 },
    { url: `${base}/categories`, lastModified: now, changeFrequency: "weekly", priority: 0.6 },
    { url: `${base}/brands`, lastModified: now, changeFrequency: "weekly", priority: 0.6 },
    { url: `${base}/search`, lastModified: now, changeFrequency: "weekly", priority: 0.5 },
  ]

  // Enumerate products by paging through public catalog.
  // Operational cap per run to keep sitemap generation bounded.
  const PER_PAGE = 200
  const MAX_PRODUCTS = 5000
  let seen = 0
  let page = 1
  let totalPages = 1

  while (page <= totalPages && seen < MAX_PRODUCTS) {
    const url = new URL(`${env.apiBaseUrl}/public/products`)
    url.searchParams.set("page", String(page))
    url.searchParams.set("per_page", String(PER_PAGE))

    const data = await fetchEnvelope<unknown>(url.toString())
    const rec = asRecord(data) ?? {}
    const items = asArray(rec.items) ?? []
    const pagination = asRecord(rec.pagination) ?? {}
    totalPages = asNumber(pagination.total_pages) ?? totalPages

    for (const item of items) {
      if (seen >= MAX_PRODUCTS) break
      const p = asRecord(item)
      const slug = asString(p?.slug)
      const id = asString(p?.id)
      const path = slug ? `/p/${encodeURIComponent(slug)}` : id ? `/products/${id}` : null
      if (!path) continue
      out.push({
        url: `${base}${path}`,
        lastModified: now,
        changeFrequency: "weekly",
        priority: 0.7,
      })
      seen++
    }

    page++
    if (items.length === 0) break
  }

  return out
}

