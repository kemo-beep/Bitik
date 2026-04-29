import { ProductDetailClient } from "@/app/(storefront)/products/product-detail-client"
import type { Metadata } from "next"
import { env } from "@/lib/env"

type Envelope<T> = { success?: boolean; data?: T }

async function fetchProductById(id: string): Promise<Record<string, unknown> | null> {
  try {
    const res = await fetch(`${env.apiBaseUrl}/public/products/${id}`, { method: "GET" })
    if (!res.ok) return null
    const json = (await res.json()) as Envelope<Record<string, unknown>>
    return json?.data ?? null
  } catch {
    return null
  }
}

function asString(value: unknown): string | null {
  return typeof value === "string" ? value : null
}

function asNumber(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>
}): Promise<Metadata> {
  const { id } = await params
  const product = await fetchProductById(id)
  const name = asString(product?.name) ?? "Product"
  const description = asString(product?.description) ?? "Product details on Bitik."
  const slug = asString(product?.slug)
  const canonicalPath = slug ? `/p/${encodeURIComponent(slug)}` : `/products/${id}`
  const images = Array.isArray(product?.images)
    ? product?.images
        .map((x) => (x && typeof x === "object" ? (x as Record<string, unknown>).url : null))
        .filter((u): u is string => typeof u === "string")
        .slice(0, 1)
    : []

  return {
    title: name,
    description,
    alternates: { canonical: `${env.oauthRedirectBaseUrl}${canonicalPath}` },
    openGraph: {
      title: name,
      description,
      url: `${env.oauthRedirectBaseUrl}${canonicalPath}`,
      type: "website",
      images: images.length ? images : undefined,
    },
    twitter: {
      card: "summary_large_image",
      title: name,
      description,
      images: images.length ? images : undefined,
    },
  }
}

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  // JSON-LD is rendered server-side so crawlers get structured data.
  // We keep the interactive UI in a client component.
  return (
    <>
      <ProductJsonLd productId={id} />
      <ProductDetailClient productId={id} />
    </>
  )
}

async function ProductJsonLd({ productId }: { productId: string }) {
  const product = await fetchProductById(productId)
  if (!product) return null

  const name = asString(product.name) ?? "Product"
  const description = asString(product.description) ?? ""
  const currency = asString(product.currency) ?? "MMK"
  const minCents = asNumber(product.min_price_cents) ?? 0
  const maxCents = asNumber(product.max_price_cents) ?? minCents
  const rating = Number(asString(product.rating) ?? "0")
  const reviewCount = asNumber(product.review_count) ?? 0
  const slug = asString(product.slug)
  const url = `${env.oauthRedirectBaseUrl}${slug ? `/p/${encodeURIComponent(slug)}` : `/products/${productId}`}`
  const images = Array.isArray(product.images)
    ? product.images
        .map((x) => (x && typeof x === "object" ? (x as Record<string, unknown>).url : null))
        .filter((u): u is string => typeof u === "string")
    : []

  const jsonLd = {
    "@context": "https://schema.org",
    "@type": "Product",
    name,
    description: description || undefined,
    image: images.length ? images : undefined,
    offers: {
      "@type": "AggregateOffer",
      priceCurrency: currency,
      lowPrice: (minCents / 100).toFixed(2),
      highPrice: (maxCents / 100).toFixed(2),
      url,
      availability: "https://schema.org/InStock",
    },
    aggregateRating:
      Number.isFinite(rating) && rating > 0
        ? {
            "@type": "AggregateRating",
            ratingValue: rating.toFixed(1),
            reviewCount,
          }
        : undefined,
  }

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
    />
  )
}
