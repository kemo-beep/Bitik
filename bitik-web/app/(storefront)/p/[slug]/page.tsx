import { ProductDetailClient } from "@/app/(storefront)/products/product-detail-client"
import type { Metadata } from "next"
import { env } from "@/lib/env"

type Envelope<T> = { success?: boolean; data?: T }

async function fetchProductBySlug(slug: string): Promise<Record<string, unknown> | null> {
  try {
    const res = await fetch(`${env.apiBaseUrl}/public/products/slug/${encodeURIComponent(slug)}`, { method: "GET" })
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

export async function generateMetadata(
  { params }: { params: { slug: string } }
): Promise<Metadata> {
  const product = await fetchProductBySlug(params.slug)
  const name = asString(product?.name) ?? "Product"
  const description = asString(product?.description) ?? "Product details on Bitik."
  const canonicalPath = `/p/${encodeURIComponent(params.slug)}`
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

export default function Page({ params }: { params: { slug: string } }) {
  return <ProductDetailClient productSlug={params.slug} />
}

