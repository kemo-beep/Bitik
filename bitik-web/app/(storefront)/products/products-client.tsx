"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Link from "next/link"
import { usePathname, useSearchParams } from "next/navigation"
import { listProducts, type ProductSort } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { ProductCard } from "@/components/storefront/product-card"
import { PagedPagination } from "@/components/storefront/paged-pagination"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { asArray, asNumber, asRecord } from "@/lib/safe"
import { Skeleton } from "@/components/ui/skeleton"

function coerceInt(value: string | null, fallback: number) {
  const n = value ? Number.parseInt(value, 10) : Number.NaN
  return Number.isFinite(n) && n > 0 ? n : fallback
}

function buildHref(pathname: string, sp: URLSearchParams, page: number) {
  const next = new URLSearchParams(sp)
  next.set("page", String(page))
  return `${pathname}?${next.toString()}`
}

export function ProductsClient({
  searchParams,
}: {
  searchParams: Record<string, string | string[] | undefined>
}) {
  const pathname = usePathname()
  const live = useSearchParams()

  // Prefer live params (client navigation), fallback to initial server props.
  const q =
    live.get("q") ?? (typeof searchParams.q === "string" ? searchParams.q : "")
  const sort = (live.get("sort") ??
    (typeof searchParams.sort === "string" ? searchParams.sort : "")) as
    | ProductSort
    | ""
  const page = coerceInt(
    live.get("page") ??
      (typeof searchParams.page === "string" ? searchParams.page : null),
    1
  )
  const perPage = coerceInt(
    live.get("per_page") ??
      (typeof searchParams.per_page === "string"
        ? searchParams.per_page
        : null),
    24
  )
  const minPrice =
    live.get("min_price_cents") ??
    (typeof searchParams.min_price_cents === "string"
      ? searchParams.min_price_cents
      : "")
  const maxPrice =
    live.get("max_price_cents") ??
    (typeof searchParams.max_price_cents === "string"
      ? searchParams.max_price_cents
      : "")

  const params = React.useMemo(() => {
    const out: Record<string, unknown> = {
      page,
      per_page: perPage,
    }
    if (q.trim() !== "") out.q = q.trim()
    if (sort) out.sort = sort
    const min = minPrice.trim() === "" ? null : Number.parseInt(minPrice, 10)
    const max = maxPrice.trim() === "" ? null : Number.parseInt(maxPrice, 10)
    if (min != null && Number.isFinite(min)) out.min_price_cents = min
    if (max != null && Number.isFinite(max)) out.max_price_cents = max
    return out
  }, [page, perPage, q, sort, minPrice, maxPrice])

  const products = useQuery({
    queryKey: queryKeys.public.listProducts(params),
    queryFn: () =>
      listProducts({
        page,
        per_page: perPage,
        q: q.trim() || undefined,
        sort: sort || undefined,
        min_price_cents: minPrice.trim()
          ? Number.parseInt(minPrice, 10)
          : undefined,
        max_price_cents: maxPrice.trim()
          ? Number.parseInt(maxPrice, 10)
          : undefined,
      }),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const data = asRecord(products.data) ?? {}
  const items = asArray(data.items) ?? []
  const pagination = asRecord(data.pagination) ?? {}
  const totalPages = asNumber(pagination.total_pages) ?? 1

  const hrefForPage = React.useCallback(
    (p: number) => buildHref(pathname, live, p),
    [pathname, live]
  )

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-6 lg:px-6">
      <header className="flex flex-col gap-4 border-b pb-5 md:flex-row md:items-end md:justify-between">
        <div className="max-w-sm">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            Products
          </h1>
          <p className="text-sm text-muted-foreground">
            Browse active listings from verified sellers.
          </p>
        </div>

        <form
          method="GET"
          className="grid gap-2 xs:grid-cols-2 md:flex md:flex-1 md:flex-wrap md:items-center md:justify-end"
        >
          <Input
            name="q"
            placeholder="Search products"
            defaultValue={q}
            className="xs:col-span-2 md:w-60 lg:w-72"
          />
          <Input
            name="min_price_cents"
            inputMode="numeric"
            placeholder="Min (cents)"
            defaultValue={minPrice}
            className="min-w-0 md:w-32"
          />
          <Input
            name="max_price_cents"
            inputMode="numeric"
            placeholder="Max (cents)"
            defaultValue={maxPrice}
            className="min-w-0 md:w-32"
          />
          <select
            name="sort"
            defaultValue={sort}
            className="h-9 min-w-0 rounded-md border bg-background px-2.5 text-sm md:w-32"
            aria-label="Sort products"
          >
            <option value="">Sort</option>
            <option value="newest">Newest</option>
            <option value="popular">Popular</option>
            <option value="rating">Rating</option>
            <option value="price_asc">Price ↑</option>
            <option value="price_desc">Price ↓</option>
          </select>
          <input type="hidden" name="page" value="1" />
          <input type="hidden" name="per_page" value={String(perPage)} />
          <Button type="submit" className="w-full xs:w-auto">
            Apply
          </Button>
          <Button
            type="button"
            variant="ghost"
            className="w-full xs:w-auto"
            onClick={() => {
              // Reset by navigating to base route.
              window.location.href = pathname
            }}
          >
            Reset
          </Button>
        </form>
      </header>

      {products.isError ? (
        <div className="mt-8 rounded-lg border bg-destructive/5 p-4 text-sm">
          Could not load products.{" "}
          <Link href={pathname} className="underline">
            Try again
          </Link>
          .
        </div>
      ) : null}

      <section className="mt-5">
        {products.isLoading ? (
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 12 }).map((_, i) => (
              <Skeleton key={i} className="aspect-[4/5] rounded-md" />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
            {items.length ? (
              items.map((p, i) => <ProductCard key={String(i)} product={p} />)
            ) : (
              <div className="col-span-full rounded-md border bg-muted/30 p-4 text-sm text-muted-foreground">
                No products found.
              </div>
            )}
          </div>
        )}
      </section>

      <PagedPagination
        page={page}
        totalPages={totalPages}
        hrefForPage={hrefForPage}
      />
    </div>
  )
}
