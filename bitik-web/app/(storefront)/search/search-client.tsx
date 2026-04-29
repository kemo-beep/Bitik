"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Link from "next/link"
import { usePathname, useSearchParams } from "next/navigation"
import { listProducts, type ProductSort } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { ProductCard } from "@/components/storefront/product-card"
import { PagedPagination } from "@/components/storefront/paged-pagination"
import { Skeleton } from "@/components/ui/skeleton"
import { asArray, asNumber, asRecord } from "@/lib/safe"
import { useAnalytics } from "@/lib/analytics"
import { analyticsEvents } from "@/lib/analytics-events"

function coerceInt(value: string | null, fallback: number) {
  const n = value ? Number.parseInt(value, 10) : Number.NaN
  return Number.isFinite(n) && n > 0 ? n : fallback
}

function buildHref(pathname: string, sp: URLSearchParams, page: number) {
  const next = new URLSearchParams(sp)
  next.set("page", String(page))
  return `${pathname}?${next.toString()}`
}

export function SearchClient({
  searchParams,
}: {
  searchParams: Record<string, string | string[] | undefined>
}) {
  const analytics = useAnalytics()

  const pathname = usePathname()
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

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-8 lg:px-6">
      <section>
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

