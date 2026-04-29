"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Link from "next/link"
import { usePathname, useSearchParams } from "next/navigation"
import { getCategory, listCategoryProducts } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { ProductCard } from "@/components/storefront/product-card"
import { PagedPagination } from "@/components/storefront/paged-pagination"
import { asArray, asNumber, asRecord, asString } from "@/lib/safe"
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

export function CategoryClient({
  categoryId,
  searchParams,
}: {
  categoryId: string
  searchParams: Record<string, string | string[] | undefined>
}) {
  const pathname = usePathname()
  const live = useSearchParams()

  const page = coerceInt(live.get("page") ?? (typeof searchParams.page === "string" ? searchParams.page : null), 1)
  const perPage = coerceInt(
    live.get("per_page") ?? (typeof searchParams.per_page === "string" ? searchParams.per_page : null),
    24
  )

  const category = useQuery({
    queryKey: queryKeys.public.category(categoryId),
    queryFn: () => getCategory(categoryId),
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    retry: 1,
  })

  const list = useQuery({
    queryKey: queryKeys.public.listCategoryProducts(categoryId, { page, per_page: perPage }),
    queryFn: () => listCategoryProducts(categoryId, { page, per_page: perPage }),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const listData = asRecord(list.data) ?? {}
  const items = asArray(listData.items) ?? []
  const pagination = asRecord(listData.pagination) ?? {}
  const totalPages = asNumber(pagination.total_pages) ?? 1

  const title = asString(asRecord(category.data)?.name) ?? "Category"

  const hrefForPage = React.useCallback(
    (p: number) => buildHref(pathname, live, p),
    [pathname, live]
  )

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-8 lg:px-6">
      <header className="flex flex-col gap-2">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">{title}</h1>
        <p className="text-sm text-muted-foreground">Products in this category.</p>
      </header>

      {list.isError ? (
        <div className="mt-8 rounded-xl border bg-destructive/5 p-6 text-sm">
          Could not load category products.{" "}
          <Link href={pathname} className="underline">
            Try again
          </Link>
          .
        </div>
      ) : null}

      <section className="mt-8">
        {list.isLoading ? (
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 12 }).map((_, i) => (
              <Skeleton key={i} className="h-32 rounded-xl" />
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
            {items.length ? (
              items.map((p, i) => <ProductCard key={String(i)} product={p} />)
            ) : (
              <div className="col-span-full rounded-xl border bg-muted/30 p-6 text-sm text-muted-foreground">
                No products found.
              </div>
            )}
          </div>
        )}
      </section>

      <PagedPagination page={page} totalPages={totalPages} hrefForPage={hrefForPage} />
    </div>
  )
}
