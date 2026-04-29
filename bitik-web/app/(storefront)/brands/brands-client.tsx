"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Image from "next/image"
import Link from "next/link"
import { usePathname, useSearchParams } from "next/navigation"
import { listBrands } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { routes } from "@/lib/routes"
import { asArray, asNumber, asRecord, asString } from "@/lib/safe"
import { Skeleton } from "@/components/ui/skeleton"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { PagedPagination } from "@/components/storefront/paged-pagination"

function coerceInt(value: string | null, fallback: number) {
  const n = value ? Number.parseInt(value, 10) : Number.NaN
  return Number.isFinite(n) && n > 0 ? n : fallback
}

function buildHref(pathname: string, sp: URLSearchParams, page: number) {
  const next = new URLSearchParams(sp)
  next.set("page", String(page))
  return `${pathname}?${next.toString()}`
}

export function BrandsClient({
  searchParams,
}: {
  searchParams: Record<string, string | string[] | undefined>
}) {
  const pathname = usePathname()
  const live = useSearchParams()

  const page = coerceInt(live.get("page") ?? (typeof searchParams.page === "string" ? searchParams.page : null), 1)
  const perPage = coerceInt(
    live.get("per_page") ?? (typeof searchParams.per_page === "string" ? searchParams.per_page : null),
    24
  )

  const q = useQuery({
    queryKey: queryKeys.public.brands({ page, per_page: perPage }),
    queryFn: () => listBrands({ page, per_page: perPage }),
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    retry: 1,
  })

  const data = asRecord(q.data) ?? {}
  const items = asArray(data.items) ?? []
  const pagination = asRecord(data.pagination) ?? {}
  const totalPages = asNumber(pagination.total_pages) ?? 1

  const hrefForPage = React.useCallback(
    (p: number) => buildHref(pathname, live, p),
    [pathname, live]
  )

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-8 lg:px-6">
      <header className="flex flex-col gap-2">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">Brands</h1>
        <p className="text-sm text-muted-foreground">Browse brands in the public catalog.</p>
      </header>

      <section className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {q.isLoading ? (
          Array.from({ length: 12 }).map((_, i) => <Skeleton key={i} className="h-32 rounded-xl" />)
        ) : q.isError ? (
          <div className="col-span-full rounded-xl border bg-destructive/5 p-6 text-sm">Could not load brands.</div>
        ) : (
          items.map((b, i) => {
            const rec = asRecord(b)
            const id = asString(rec?.id)
            const name = asString(rec?.name) ?? "Unnamed brand"
            const logoUrl = asString(rec?.logo_url)
            const href = id ? routes.storefront.brand(id) : routes.storefront.brands
            return (
              <Card key={id ?? String(i)} className="h-full">
                <CardHeader>
                  <CardTitle className="line-clamp-1">
                    <Link href={href} className="hover:underline">
                      {name}
                    </Link>
                  </CardTitle>
                </CardHeader>
                <CardContent className="flex items-center gap-3">
                  {logoUrl ? (
                    <Image src={logoUrl} alt="" width={48} height={48} className="size-12 rounded-md object-contain" />
                  ) : (
                    <div className="size-12 rounded-md border bg-muted" />
                  )}
                  <div className="text-xs text-muted-foreground">View products →</div>
                </CardContent>
              </Card>
            )
          })
        )}
      </section>

      <PagedPagination page={page} totalPages={totalPages} hrefForPage={hrefForPage} />
    </div>
  )
}

