"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Link from "next/link"
import { useParams } from "next/navigation"
import { getProduct, getProductBySlug, getSeller, listProductReviews, listRelatedProducts } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asNumber, asRecord, asString } from "@/lib/safe"
import { ImageGallery, type GalleryImage } from "@/components/ui/image-gallery"
import { ProductCard } from "@/components/storefront/product-card"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatDate, formatMoney, formatMoneyRange, formatNumber } from "@/lib/format"
import { routes } from "@/lib/routes"
import { Skeleton } from "@/components/ui/skeleton"
import { useAnalytics, analyticsEvents } from "@/lib/analytics"

function productToGalleryImages(p: Record<string, unknown>): GalleryImage[] {
  const images = asArray(p.images) ?? []
  const out: GalleryImage[] = []
  for (const img of images) {
    const rec = asRecord(img)
    const url = asString(rec?.url)
    if (!url) continue
    out.push({
      src: url,
      alt: asString(rec?.alt_text) ?? asString(p.name) ?? "Product image",
    })
  }
  return out
}

export function ProductDetailClient({
  productId,
  productSlug,
}: {
  productId?: string
  productSlug?: string
}) {
  const routeParams = useParams<{ id?: string; slug?: string }>()
  const effectiveProductId = productId ?? routeParams?.id
  const effectiveProductSlug = productSlug ?? routeParams?.slug

  const product = useQuery({
    queryKey: effectiveProductId
      ? queryKeys.public.product(effectiveProductId)
      : queryKeys.public.productSlug(effectiveProductSlug ?? ""),
    queryFn: () => {
      if (effectiveProductId) return getProduct(effectiveProductId)
      if (effectiveProductSlug) return getProductBySlug(effectiveProductSlug)
      return Promise.resolve({})
    },
    enabled: Boolean(effectiveProductId || effectiveProductSlug),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const p = asRecord(product.data) ?? {}
  const analytics = useAnalytics()
  const id = asString(p.id) ?? effectiveProductId ?? ""
  const name = asString(p.name) ?? "Product"
  const description = asString(p.description) ?? ""
  const currency = asString(p.currency) ?? "MMK"
  const minCents = asNumber(p.min_price_cents) ?? 0
  const maxCents = asNumber(p.max_price_cents) ?? minCents
  const rating = Number(asString(p.rating) ?? "0")
  const reviewCount = asNumber(p.review_count) ?? 0
  const sellerId = asString(p.seller_id) ?? ""

  const variants = asArray(p.variants) ?? []
  const [variantIdx, setVariantIdx] = React.useState(0)
  const variant = asRecord(variants[Math.min(variantIdx, Math.max(0, variants.length - 1))]) ?? null
  const variantName = asString(variant?.name) ?? asString(variant?.sku) ?? ""
  const variantPriceCents = asNumber(variant?.price_cents)

  const reviews = useQuery({
    queryKey: queryKeys.public.productReviews(id),
    queryFn: () => listProductReviews(id),
    enabled: Boolean(id),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const related = useQuery({
    queryKey: queryKeys.public.productRelated(id),
    queryFn: () => listRelatedProducts(id),
    enabled: Boolean(id),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const seller = useQuery({
    queryKey: queryKeys.public.seller(sellerId),
    queryFn: () => getSeller(sellerId),
    enabled: Boolean(sellerId),
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    retry: 1,
  })

  const gallery = productToGalleryImages(p)
  const reviewItems = asArray(asRecord(reviews.data)?.items) ?? []
  const relatedItems = asArray(related.data) ?? []

  const sellerData = asRecord(seller.data) ?? {}
  React.useEffect(() => {
    if (!id) return
    analytics.track({
      name: analyticsEvents.productView,
      properties: { product_id: id, seller_id: sellerId },
    })
  }, [analytics, id, sellerId])

  const shopName = asString(sellerData.shop_name) ?? "Seller"
  const shopHref = sellerId ? routes.storefront.seller(sellerId) : routes.storefront.products

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-8 lg:px-6">
      {product.isLoading ? (
        <div className="grid gap-8 lg:grid-cols-2">
          <Skeleton className="aspect-square w-full rounded-2xl" />
          <div className="flex flex-col gap-3">
            <Skeleton className="h-8 w-2/3" />
            <Skeleton className="h-5 w-1/3" />
            <Skeleton className="h-24 w-full" />
            <Skeleton className="h-10 w-40" />
          </div>
        </div>
      ) : product.isError ? (
        <div className="rounded-xl border bg-destructive/5 p-6 text-sm">Could not load product.</div>
      ) : (
        <>
          <section className="grid gap-8 lg:grid-cols-2">
            <ImageGallery images={gallery} />

            <div className="flex flex-col gap-4">
              <header className="flex flex-col gap-1">
                <h1 className="font-heading text-2xl font-semibold tracking-tight md:text-3xl">{name}</h1>
                <div className="text-sm text-muted-foreground">
                  {Number.isFinite(rating) && rating > 0 ? `${rating.toFixed(1)}★` : "No rating"} ·{" "}
                  {formatNumber(reviewCount)} reviews
                </div>
              </header>

              <div className="text-lg font-semibold">
                {variantPriceCents != null
                  ? formatMoney(variantPriceCents / 100, { currency })
                  : formatMoneyRange(minCents / 100, maxCents / 100, { currency })}
              </div>

              {variants.length ? (
                <div className="flex flex-col gap-2">
                  <div className="text-sm font-medium">Variant</div>
                  <select
                    className="h-10 rounded-md border bg-background px-3 text-sm"
                    value={String(variantIdx)}
                    onChange={(e) => setVariantIdx(Number.parseInt(e.target.value, 10))}
                    aria-label="Select variant"
                  >
                    {variants.map((v, i) => {
                      const rec = asRecord(v) ?? {}
                      const label = asString(rec.name) ?? asString(rec.sku) ?? `Variant ${i + 1}`
                      const cents = asNumber(rec.price_cents)
                      return (
                        <option key={String(i)} value={String(i)}>
                          {label}
                          {cents != null ? ` · ${formatMoney(cents / 100, { currency: asString(rec.currency) ?? currency })}` : ""}
                        </option>
                      )
                    })}
                  </select>
                  {variantName ? <div className="text-xs text-muted-foreground">Selected: {variantName}</div> : null}
                </div>
              ) : null}

              <div className="flex flex-col gap-2 sm:flex-row">
                <Button
                  title="Cart API not in Phase 3"
                  onClick={() =>
                    analytics.track({
                      name: analyticsEvents.addToCart,
                      properties: { product_id: id, variant: variantName || null },
                    })
                  }
                >
                  Add to cart (stub)
                </Button>
                <Button disabled variant="outline" title="Wishlist API not in Phase 3">
                  Wishlist (stub)
                </Button>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle>Seller</CardTitle>
                </CardHeader>
                <CardContent className="flex items-center justify-between gap-3">
                  <div>
                    <div className="font-medium">{shopName}</div>
                    <div className="text-xs text-muted-foreground">Ships fast · Secure payments</div>
                  </div>
                  <Button
                    variant="outline"
                    nativeButton={false}
                    render={<Link href={shopHref}>Visit shop</Link>}
                  />
                </CardContent>
              </Card>

              {description ? (
                <section className="rounded-xl border bg-muted/20 p-4">
                  <h2 className="text-sm font-medium">About</h2>
                  <p className="mt-2 whitespace-pre-wrap text-sm text-muted-foreground">{description}</p>
                </section>
              ) : null}

              <section className="rounded-xl border p-4">
                <h2 className="text-sm font-medium">Shipping & payment</h2>
                <ul className="mt-2 list-disc pl-5 text-sm text-muted-foreground">
                  <li>Delivery estimates shown at checkout (stub).</li>
                  <li>Cashless payments supported (stub).</li>
                  <li>Returns and refunds policy (stub).</li>
                </ul>
              </section>
            </div>
          </section>

          <section className="mt-12">
            <h2 className="font-heading text-xl font-semibold tracking-tight">Reviews</h2>
            <div className="mt-4 grid gap-3 md:grid-cols-2">
              {reviews.isLoading ? (
                Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-24 rounded-xl" />)
              ) : reviews.isError ? (
                <div className="col-span-full rounded-xl border bg-destructive/5 p-6 text-sm">
                  Could not load reviews.
                </div>
              ) : reviewItems.length ? (
                reviewItems.slice(0, 6).map((r, i) => {
                  const rec = asRecord(r) ?? {}
                  const rr = asNumber(rec.rating) ?? 0
                  const title = asString(rec.title) ?? "Review"
                  const body = asString(rec.body) ?? ""
                  const rawCreatedAt = rec.created_at
                  const createdAt =
                    typeof rawCreatedAt === "string" || typeof rawCreatedAt === "number"
                      ? rawCreatedAt
                      : rawCreatedAt instanceof Date
                        ? rawCreatedAt
                        : null
                  return (
                    <Card key={String(i)}>
                      <CardHeader>
                        <CardTitle className="text-base">
                          {rr ? `${rr}★ · ` : ""}
                          {title}
                        </CardTitle>
                      </CardHeader>
                      <CardContent className="text-sm text-muted-foreground">
                        <div className="line-clamp-5">{body || "—"}</div>
                        {createdAt != null ? (
                          <div className="mt-2 text-xs text-muted-foreground">{formatDate(createdAt)}</div>
                        ) : null}
                      </CardContent>
                    </Card>
                  )
                })
              ) : (
                <div className="col-span-full rounded-xl border bg-muted/30 p-6 text-sm text-muted-foreground">
                  No reviews yet.
                </div>
              )}
            </div>
          </section>

          <section className="mt-12">
            <div className="flex items-end justify-between gap-4">
              <h2 className="font-heading text-xl font-semibold tracking-tight">Related products</h2>
              <Link href={routes.storefront.products} className="text-sm underline underline-offset-4">
                Browse more
              </Link>
            </div>
            <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {related.isLoading ? (
                Array.from({ length: 4 }).map((_, i) => <Skeleton key={i} className="h-32 rounded-xl" />)
              ) : related.isError ? (
                <div className="col-span-full rounded-xl border bg-destructive/5 p-6 text-sm">
                  Could not load related products.
                </div>
              ) : relatedItems.length ? (
                relatedItems.slice(0, 8).map((rp, i) => <ProductCard key={String(i)} product={rp} />)
              ) : (
                <div className="col-span-full rounded-xl border bg-muted/30 p-6 text-sm text-muted-foreground">
                  No related products.
                </div>
              )}
            </div>
          </section>
        </>
      )}
    </div>
  )
}

