"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Link from "next/link"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { listProducts, type ProductSort } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { ProductCard } from "@/components/storefront/product-card"
import { PagedPagination } from "@/components/storefront/paged-pagination"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { asArray, asNumber, asRecord } from "@/lib/safe"
import { useFeatureFlags } from "@/lib/feature-flags"
import { useAnalytics } from "@/lib/analytics"
import { analyticsEvents } from "@/lib/analytics-events"

const RECENTS_KEY = "bitik_recent_searches_v1"

function coerceInt(value: string | null, fallback: number) {
  const n = value ? Number.parseInt(value, 10) : Number.NaN
  return Number.isFinite(n) && n > 0 ? n : fallback
}

function buildHref(pathname: string, sp: URLSearchParams, page: number) {
  const next = new URLSearchParams(sp)
  next.set("page", String(page))
  return `${pathname}?${next.toString()}`
}

function readRecents(): string[] {
  try {
    const raw = localStorage.getItem(RECENTS_KEY)
    const parsed = raw ? (JSON.parse(raw) as unknown) : []
    return Array.isArray(parsed) ? parsed.filter((x) => typeof x === "string") : []
  } catch {
    return []
  }
}

function writeRecents(next: string[]) {
  try {
    localStorage.setItem(RECENTS_KEY, JSON.stringify(next.slice(0, 8)))
  } catch {
    // ignore
  }
}

export function SearchClient({
  searchParams,
}: {
  searchParams: Record<string, string | string[] | undefined>
}) {
  const flags = useFeatureFlags()
  const analytics = useAnalytics()

  const pathname = usePathname()
  const router = useRouter()
  const live = useSearchParams()

  const q = live.get("q") ?? (typeof searchParams.q === "string" ? searchParams.q : "")
  const sort = (live.get("sort") ??
    (typeof searchParams.sort === "string" ? searchParams.sort : "")) as ProductSort | ""
  const page = coerceInt(live.get("page") ?? (typeof searchParams.page === "string" ? searchParams.page : null), 1)
  const perPage = coerceInt(
    live.get("per_page") ?? (typeof searchParams.per_page === "string" ? searchParams.per_page : null),
    24
  )

  const categoryId = live.get("category_id") ?? (typeof searchParams.category_id === "string" ? searchParams.category_id : "")
  const brandId = live.get("brand_id") ?? (typeof searchParams.brand_id === "string" ? searchParams.brand_id : "")
  const sellerId = live.get("seller_id") ?? (typeof searchParams.seller_id === "string" ? searchParams.seller_id : "")
  const minPrice = live.get("min_price_cents") ?? (typeof searchParams.min_price_cents === "string" ? searchParams.min_price_cents : "")
  const maxPrice = live.get("max_price_cents") ?? (typeof searchParams.max_price_cents === "string" ? searchParams.max_price_cents : "")

  const params = React.useMemo(() => {
    const out: Record<string, unknown> = { page, per_page: perPage }
    if (q.trim()) out.q = q.trim()
    if (sort) out.sort = sort
    if (categoryId.trim()) out.category_id = categoryId.trim()
    if (brandId.trim()) out.brand_id = brandId.trim()
    if (sellerId.trim()) out.seller_id = sellerId.trim()
    const min = minPrice.trim() ? Number.parseInt(minPrice, 10) : null
    const max = maxPrice.trim() ? Number.parseInt(maxPrice, 10) : null
    if (min != null && Number.isFinite(min)) out.min_price_cents = min
    if (max != null && Number.isFinite(max)) out.max_price_cents = max
    return out
  }, [page, perPage, q, sort, categoryId, brandId, sellerId, minPrice, maxPrice])

  const results = useQuery({
    queryKey: queryKeys.public.listProducts(params),
    queryFn: () =>
      listProducts({
        page,
        per_page: perPage,
        q: q.trim() || undefined,
        sort: sort || undefined,
        category_id: categoryId.trim() || undefined,
        brand_id: brandId.trim() || undefined,
        seller_id: sellerId.trim() || undefined,
        min_price_cents: minPrice.trim() ? Number.parseInt(minPrice, 10) : undefined,
        max_price_cents: maxPrice.trim() ? Number.parseInt(maxPrice, 10) : undefined,
      }),
    enabled: q.trim().length > 0 || categoryId.trim() !== "" || brandId.trim() !== "" || sellerId.trim() !== "",
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })
  const searchEnabled =
    q.trim().length > 0 ||
    categoryId.trim() !== "" ||
    brandId.trim() !== "" ||
    sellerId.trim() !== ""

  const data = asRecord(results.data) ?? {}
  const items = asArray(data.items) ?? []
  const pagination = asRecord(data.pagination) ?? {}
  const totalPages = asNumber(pagination.total_pages) ?? 1

  const hrefForPage = React.useCallback(
    (p: number) => buildHref(pathname, live, p),
    [pathname, live]
  )

  const [recents, setRecents] = React.useState<string[]>([])
  const [draftQ, setDraftQ] = React.useState(q)
  React.useEffect(() => {
    setDraftQ(q)
  }, [q])
  React.useEffect(() => {
    const handle = window.setTimeout(() => {
      const next = new URLSearchParams(live.toString())
      if (draftQ.trim()) next.set("q", draftQ.trim())
      else next.delete("q")
      next.set("page", "1")
      router.replace(`${pathname}?${next.toString()}`)
    }, 350)
    return () => window.clearTimeout(handle)
  }, [draftQ, live, pathname, router])
  React.useEffect(() => {
    setRecents(readRecents())
  }, [])

  function onSubmitted(nextQ: string) {
    const trimmed = nextQ.trim()
    if (!trimmed) return
    const next = [trimmed, ...recents.filter((x) => x !== trimmed)]
    setRecents(next)
    writeRecents(next)
  }

  const showStubs = Boolean(flags.search_suggestions_stub || flags.search_trending_stub)

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-8 lg:px-6">
      <header className="flex flex-col gap-2">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">Search</h1>
        <p className="text-sm text-muted-foreground">Search products with filters and pagination.</p>
      </header>

      <section className="mt-6 rounded-2xl border p-4">
        <form
          method="GET"
          className="grid gap-2 md:grid-cols-6"
          onSubmit={(e) => {
            const form = e.currentTarget
            const fd = new FormData(form)
            const nextQ = String(fd.get("q") ?? "")
            onSubmitted(nextQ)
            analytics.track({ name: analyticsEvents.searchSubmit, properties: { q: nextQ.trim() } })
          }}
        >
            <Input
              name="q"
              placeholder="Search products"
              value={draftQ}
              onChange={(e) => setDraftQ(e.target.value)}
              className="md:col-span-2"
            />
          <Input name="category_id" placeholder="Category id" defaultValue={categoryId} />
          <Input name="brand_id" placeholder="Brand id" defaultValue={brandId} />
          <Input name="seller_id" placeholder="Seller id" defaultValue={sellerId} />
          <select
            name="sort"
            defaultValue={sort}
            className="h-10 rounded-md border bg-background px-3 text-sm"
            aria-label="Sort results"
          >
            <option value="">Sort</option>
            <option value="newest">Newest</option>
            <option value="popular">Popular</option>
            <option value="rating">Rating</option>
            <option value="price_asc">Price ↑</option>
            <option value="price_desc">Price ↓</option>
          </select>
          <Input name="min_price_cents" inputMode="numeric" placeholder="Min (cents)" defaultValue={minPrice} />
          <Input name="max_price_cents" inputMode="numeric" placeholder="Max (cents)" defaultValue={maxPrice} />
          <input type="hidden" name="page" value="1" />
          <input type="hidden" name="per_page" value={String(perPage)} />
          <div className="flex gap-2 md:col-span-6">
            <Button type="submit">Search</Button>
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                window.location.href = pathname
              }}
            >
              Reset
            </Button>
          </div>
        </form>

        {recents.length ? (
          <div className="mt-3 flex flex-wrap gap-2">
            <span className="text-xs text-muted-foreground">Recent:</span>
            {recents.map((r) => (
              <Link
                key={r}
                href={`${pathname}?q=${encodeURIComponent(r)}`}
                className="rounded-full border px-2 py-0.5 text-xs hover:bg-muted"
              >
                {r}
              </Link>
            ))}
          </div>
        ) : null}

        {showStubs ? (
          <div className="mt-4 grid gap-3 md:grid-cols-2">
            <div className="rounded-xl border bg-muted/20 p-4 text-sm">
              <div className="font-medium">Suggestions (stub)</div>
              <div className="mt-1 text-muted-foreground">
                Backend suggestions endpoint isn’t in Phase 3; this is UI-only until Phase 8.
              </div>
            </div>
            <div className="rounded-xl border bg-muted/20 p-4 text-sm">
              <div className="font-medium">Trending (stub)</div>
              <div className="mt-1 text-muted-foreground">
                Hook this up to a backend endpoint later; for now it’s feature-flagged.
              </div>
            </div>
          </div>
        ) : null}
      </section>

      <section className="mt-8">
        {!searchEnabled ? (
          <div className="rounded-xl border bg-muted/30 p-6 text-sm text-muted-foreground">
            Enter a query or filter to start searching.
          </div>
        ) : results.isLoading ? (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 12 }).map((_, i) => (
              <Skeleton key={i} className="h-32 rounded-xl" />
            ))}
          </div>
        ) : results.isError ? (
          <div className="rounded-xl border bg-destructive/5 p-6 text-sm">
            Could not load search results.{" "}
            <Link href={pathname} className="underline">
              Try again
            </Link>
            .
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {items.length ? (
              items.map((p, i) => (
                <div
                  key={String(i)}
                  onClick={() => {
                    analytics.track({
                      name: analyticsEvents.searchClick,
                      properties: { q: q.trim(), index: i },
                    })
                  }}
                >
                  <ProductCard product={p} />
                </div>
              ))
            ) : (
              <div className="col-span-full rounded-xl border bg-muted/30 p-6 text-sm text-muted-foreground">
                No results.
              </div>
            )}
          </div>
        )}
      </section>

      <PagedPagination page={page} totalPages={totalPages} hrefForPage={hrefForPage} />
    </div>
  )
}

