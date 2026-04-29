import Link from "next/link"
import Image from "next/image"
import { routes } from "@/lib/routes"
import { formatMoneyRange, formatNumber } from "@/lib/format"
import { asArray, asNumber, asRecord, asString } from "@/lib/safe"

export function ProductCard({ product }: { product: unknown }) {
  const p = asRecord(product)
  const id = asString(p?.id)
  const slug = asString(p?.slug)
  const name = asString(p?.name) ?? "Unnamed product"
  const currency = asString(p?.currency) ?? "MMK"
  const minCents = asNumber(p?.min_price_cents) ?? 0
  const maxCents = asNumber(p?.max_price_cents) ?? minCents
  const rating = Number(asString(p?.rating) ?? "0")
  const reviewCount = asNumber(p?.review_count) ?? 0
  const sold = asNumber(p?.total_sold) ?? 0
  const firstImage = asRecord(asArray(p?.images)?.[0])
  const imageUrl =
    asString(p?.primary_image_url) ??
    asString(p?.image_url) ??
    asString(firstImage?.url)

  const href = slug
    ? routes.storefront.productSlug(slug)
    : id
      ? routes.storefront.product(id)
      : routes.storefront.products

  return (
    <Link
      href={href}
      className="group grid h-full grid-rows-[auto_1fr] overflow-hidden rounded-md border bg-card transition-colors hover:border-foreground/25 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <div className="relative aspect-[4/5] overflow-hidden bg-muted">
        {imageUrl ? (
          <Image
            src={imageUrl}
            alt={name}
            fill
            sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
            className="object-cover transition-transform duration-300 group-hover:scale-[1.03]"
          />
        ) : (
          <div className="absolute inset-0 bg-[linear-gradient(135deg,var(--muted)_0%,var(--background)_100%)]" />
        )}
        <div className="absolute top-2 left-2 rounded-sm bg-background/90 px-1.5 py-0.5 text-[11px] font-medium">
          {Number.isFinite(rating) && rating > 0
            ? `${rating.toFixed(1)} ★`
            : "New"}
        </div>
      </div>
      <div className="flex min-h-24 flex-col gap-1.5 p-3">
        <div className="line-clamp-2 text-sm leading-snug font-medium">
          {name}
        </div>
        <div className="mt-auto font-semibold">
          {formatMoneyRange(minCents / 100, maxCents / 100, { currency })}
        </div>
        <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
          <span>{formatNumber(reviewCount)} reviews</span>
          <span>{formatNumber(sold)} sold</span>
        </div>
      </div>
    </Link>
  )
}
