"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export function SellerProfileClient() {
  const qc = useQueryClient()
  const seller = useQuery({ queryKey: queryKeys.seller.me(), queryFn: sellerApi.getMe })
  const [shopName, setShopName] = useState("")
  const [description, setDescription] = useState("")
  const [logoUrl, setLogoUrl] = useState("")
  const [bannerUrl, setBannerUrl] = useState("")

  const patchProfile = useMutation({
    mutationFn: () => sellerApi.patchMe({ shop_name: shopName.trim(), description: description.trim() }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.seller.me() }),
  })
  const patchMedia = useMutation({
    mutationFn: () => sellerApi.patchMedia({ logo_url: logoUrl.trim(), banner_url: bannerUrl.trim() }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.seller.me() }),
  })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Shop profile</h1>
      <SellerSectionCard title="Shop info">
        <div className="grid gap-2 md:grid-cols-2">
          <Input placeholder="Shop name" value={shopName} onChange={(e) => setShopName(e.target.value)} />
          <Input placeholder="Description" value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>
        <Button className="mt-3" onClick={() => patchProfile.mutate()} disabled={patchProfile.isPending}>
          Save profile
        </Button>
      </SellerSectionCard>
      <SellerSectionCard title="Branding media">
        <div className="grid gap-2 md:grid-cols-2">
          <Input placeholder="Logo URL" value={logoUrl} onChange={(e) => setLogoUrl(e.target.value)} />
          <Input placeholder="Banner URL" value={bannerUrl} onChange={(e) => setBannerUrl(e.target.value)} />
        </div>
        <Button className="mt-3" onClick={() => patchMedia.mutate()} disabled={patchMedia.isPending}>
          Save media
        </Button>
      </SellerSectionCard>
      <SellerJsonView value={seller.data} />
    </div>
  )
}

