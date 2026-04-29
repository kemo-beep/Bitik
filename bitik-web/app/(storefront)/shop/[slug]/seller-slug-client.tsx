"use client"

import { useQuery } from "@tanstack/react-query"
import { getSellerBySlug } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { asRecord, asString } from "@/lib/safe"
import { SellerClient } from "@/app/(storefront)/sellers/[id]/seller-client"

export function SellerSlugClient({ slug }: { slug: string }) {
  const seller = useQuery({
    queryKey: queryKeys.public.seller(`slug:${slug}`),
    queryFn: () => getSellerBySlug(slug),
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const id = asString(asRecord(seller.data)?.id)
  if (seller.isLoading) return <div className="mx-auto max-w-screen-2xl px-4 py-8">Loading…</div>
  if (seller.isError || !id) return <div className="mx-auto max-w-screen-2xl px-4 py-8">Seller not found.</div>
  return <SellerClient sellerId={id} searchParams={{}} />
}

