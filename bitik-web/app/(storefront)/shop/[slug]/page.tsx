import type { Metadata } from "next"
import { env } from "@/lib/env"
import { SellerSlugClient } from "@/app/(storefront)/shop/[slug]/seller-slug-client"

type Envelope<T> = { success?: boolean; data?: T }

async function fetchSellerBySlug(slug: string): Promise<Record<string, unknown> | null> {
  try {
    const res = await fetch(`${env.apiBaseUrl}/public/sellers/slug/${encodeURIComponent(slug)}`, { method: "GET" })
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

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>
}): Promise<Metadata> {
  const { slug } = await params
  const seller = await fetchSellerBySlug(slug)
  const shopName = asString(seller?.shop_name) ?? "Seller"
  const description = asString(seller?.description) ?? `Shop ${shopName} on Bitik.`
  const canonicalPath = `/shop/${encodeURIComponent(slug)}`
  const banner = asString(seller?.banner_url)
  const logo = asString(seller?.logo_url)
  const images = [banner, logo].filter((x): x is string => typeof x === "string" && x.trim() !== "").slice(0, 1)

  return {
    title: shopName,
    description,
    alternates: { canonical: `${env.oauthRedirectBaseUrl}${canonicalPath}` },
    openGraph: {
      title: shopName,
      description,
      url: `${env.oauthRedirectBaseUrl}${canonicalPath}`,
      type: "profile",
      images: images.length ? images : undefined,
    },
    twitter: {
      card: "summary_large_image",
      title: shopName,
      description,
      images: images.length ? images : undefined,
    },
  }
}

export default async function Page({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params
  return <SellerSlugClient slug={slug} />
}

