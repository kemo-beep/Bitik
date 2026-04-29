"use client"

import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { SellerSectionCard, SellerJsonView } from "@/components/seller/section-card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export function SellerApplyClient() {
  const [shopName, setShopName] = useState("")
  const [slug, setSlug] = useState("")
  const apply = useMutation({
    mutationFn: () => sellerApi.apply({ shop_name: shopName.trim(), slug: slug.trim(), submit: true }),
  })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Apply as seller</h1>
      <SellerSectionCard title="Seller application" description="Submit application and move to review.">
        <div className="grid gap-2 md:grid-cols-2">
          <Input placeholder="Shop name" value={shopName} onChange={(e) => setShopName(e.target.value)} />
          <Input placeholder="Slug (optional)" value={slug} onChange={(e) => setSlug(e.target.value)} />
        </div>
        <Button className="mt-3" onClick={() => apply.mutate()} disabled={!shopName.trim() || apply.isPending}>
          Submit application
        </Button>
      </SellerSectionCard>
      {apply.data ? <SellerJsonView value={apply.data} /> : null}
    </div>
  )
}

