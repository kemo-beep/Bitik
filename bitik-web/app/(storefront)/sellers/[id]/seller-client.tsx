"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Image from "next/image"
import Link from "next/link"
import { usePathname, useSearchParams } from "next/navigation"
import { getSeller, listSellerProducts, listSellerReviews } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asNumber, asRecord, asString } from "@/lib/safe"
import { ProductCard } from "@/components/storefront/product-card"
import { PagedPagination } from "@/components/storefront/paged-pagination"
import { Skeleton } from "@/components/ui/skeleton"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatNumber } from "@/lib/format"

function coerceInt(value: string | null, fallback: number) {
  const n = value ? Number.parseInt(value, 10) : Number.NaN
  return Number.isFinite(n) && n > 0 ? n : fallback
}

function buildHref(pathname: string, sp: URLSearchParams, page: number) {
  const next = new URLSearchParams(sp)
  next.set("page", String(page))
  return `${pathname}?${next.toString()}`
}

export function SellerClient({
  sellerId,
  searchParams,
}: {
  sellerId: string
  searchParams: Record<string, string | string[] | undefined>
}) {
  const pathname = usePathname()
  const live = useSearchParams()

  const page = coerceInt(live.get("page") ?? (typeof searchParams.page === "string" ? searchParams.page : null), 1)
  const perPage = coerceInt(
    live.get("per_page") ?? (typeof searchParams.per_page === "string" ? searchParams.per_page : null),
    24
  )

  const seller = useQuery({
    queryKey: queryKeys.public.seller(sellerId),
    queryFn: () => getSeller(sellerId),
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    retry: 1,
  })

  const products = useQuery({
    queryKey: queryKeys.public.listSellerProducts(sellerId, { page, per_page: perPage }),
    queryFn: () => listSellerProducts(sellerId, { page, per_page: perPage }),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const reviews = useQuery({
    queryKey: queryKeys.public.listSellerReviews(sellerId),
    queryFn: () => listSellerReviews(sellerId),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const sellerData = asRecord(seller.data) ?? {}
  const shopName = asString(sellerData.shop_name) ?? "Seller"
  const description = asString(sellerData.description) ?? ""
  const bannerUrl = asString(sellerData.banner_url)
  const logoUrl = asString(sellerData.logo_url)
  const rating = Number(asString(sellerData.rating) ?? "0")
  const totalSales = asNumber(sellerData.total_sales) ?? 0

  const listData = asRecord(products.data) ?? {}
  const items = asArray(listData.items) ?? []
  const pagination = asRecord(listData.pagination) ?? {}
  const totalPages = asNumber(pagination.total_pages) ?? 1

  const hrefForPage = React.useCallback(
    (p: number) => buildHref(pathname, live, p),
    [pathname, live]
  )

  const reviewItems = asArray(asRecord(reviews.data)?.items) ?? []

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-8 lg:px-6">
      <header className="overflow-hidden rounded-2xl border">
        {bannerUrl ? (
          <Image src={bannerUrl} alt="" width={1600} height={500} className="h-40 w-full object-cover md:h-56" />
        ) : (
          <div className="h-40 bg-muted md:h-56" />
        )}
        <div className="-mt-8 flex flex-col gap-3 px-4 pb-6 md:flex-row md:items-end md:justify-between md:px-6">
          <div className="flex items-end gap-3">
            {logoUrl ? (
              <Image
                src={logoUrl}
                alt=""
                width={96}
                height={96}
                className="size-16 rounded-xl border bg-background object-cover md:size-20"
              />
            ) : (
              <div className="size-16 rounded-xl border bg-background md:size-20" />
            )}
            <div className="pb-1">
              <h1 className="font-heading text-2xl font-semibold tracking-tight">{shopName}</h1>
              <div className="text-xs text-muted-foreground">
                {Number.isFinite(rating) && rating > 0 ? `${rating.toFixed(1)}★` : "No rating"} ·{" "}
                {formatNumber(totalSales)} sales
              </div>
            </div>
          </div>
        </div>
      </header>

      {description ? <p className="mt-6 max-w-3xl text-sm text-muted-foreground">{description}</p> : null}

      <section className="mt-10">
        <div className="flex items-end justify-between gap-4">
          <div>
            <h2 className="font-heading text-xl font-semibold tracking-tight">Products</h2>
            <p className="text-sm text-muted-foreground">All products from this seller.</p>
          </div>
        </div>

        {products.isLoading ? (
          <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-32 rounded-xl" />
            ))}
          </div>
        ) : products.isError ? (
          <div className="mt-4 rounded-xl border bg-destructive/5 p-6 text-sm">
            Could not load seller products.{" "}
            <Link href={pathname} className="underline">
              Try again
            </Link>
            .
          </div>
        ) : (
          <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
            {items.length ? (
              items.map((p, i) => <ProductCard key={String(i)} product={p} />)
            ) : (
              <div className="col-span-full rounded-xl border bg-muted/30 p-6 text-sm text-muted-foreground">
                No products found.
              </div>
            )}
          </div>
        )}

        <PagedPagination page={page} totalPages={totalPages} hrefForPage={hrefForPage} />
      </section>

      <section className="mt-10">
        <div className="flex items-end justify-between gap-4">
          <div>
            <h2 className="font-heading text-xl font-semibold tracking-tight">Reviews</h2>
            <p className="text-sm text-muted-foreground">Recent customer feedback.</p>
          </div>
        </div>

        {reviews.isLoading ? (
          <div className="mt-4 grid gap-3 md:grid-cols-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-24 rounded-xl" />
            ))}
          </div>
        ) : reviews.isError ? (
          <div className="mt-4 rounded-xl border bg-destructive/5 p-6 text-sm">Could not load reviews.</div>
        ) : (
          <div className="mt-4 grid gap-3 md:grid-cols-2">
            {reviewItems.length ? (
              reviewItems.slice(0, 6).map((r, i) => {
                const rec = asRecord(r) ?? {}
                const title = asString(rec.title) ?? "Review"
                const body = asString(rec.body) ?? ""
                const rr = asNumber(rec.rating) ?? 0
                return (
                  <Card key={String(i)}>
                    <CardHeader>
                      <CardTitle className="text-base">
                        {rr ? `${rr}★ · ` : ""}
                        {title}
                      </CardTitle>
                    </CardHeader>
                    <CardContent className="text-sm text-muted-foreground line-clamp-4">{body || "—"}</CardContent>
                  </Card>
                )
              })
            ) : (
              <div className="col-span-full rounded-xl border bg-muted/30 p-6 text-sm text-muted-foreground">
                No reviews yet.
              </div>
            )}
          </div>
        )}
      </section>
    </div>
  )
}
