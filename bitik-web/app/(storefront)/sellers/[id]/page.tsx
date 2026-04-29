import { SellerClient } from "@/app/(storefront)/sellers/[id]/seller-client"
import type { Metadata } from "next"
import { env } from "@/lib/env"

type Envelope<T> = { success?: boolean; data?: T }

async function fetchSellerById(id: string): Promise<Record<string, unknown> | null> {
  try {
    const res = await fetch(`${env.apiBaseUrl}/public/sellers/${id}`, { method: "GET" })
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
  { params }: { params: { id: string } }
): Promise<Metadata> {
  const seller = await fetchSellerById(params.id)
  const shopName = asString(seller?.shop_name) ?? "Seller"
  const description = asString(seller?.description) ?? `Shop ${shopName} on Bitik.`
  const slug = asString(seller?.slug)
  const canonicalPath = slug ? `/shop/${encodeURIComponent(slug)}` : `/sellers/${params.id}`
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

export default function Page({
  params,
  searchParams,
}: {
  params: { id: string }
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <SellerClient sellerId={params.id} searchParams={searchParams ?? {}} />
}
