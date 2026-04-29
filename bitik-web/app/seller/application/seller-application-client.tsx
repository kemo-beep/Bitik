"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export function SellerApplicationClient() {
  const qc = useQueryClient()
  const app = useQuery({ queryKey: queryKeys.seller.application(), queryFn: sellerApi.getApplication })
  const [shopName, setShopName] = useState("")
  const [submit, setSubmit] = useState(false)
  const patch = useMutation({
    mutationFn: () => sellerApi.patchApplication({ shop_name: shopName.trim() || undefined, submit }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.seller.application() }),
  })
  const createDoc = useMutation({
    mutationFn: () =>
      sellerApi.createDocument({
        document_type: "business_registration",
        file_url: "https://example.com/document.pdf",
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.seller.application() }),
  })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Application status</h1>
      <SellerSectionCard title="Edit pending application">
        <div className="flex gap-2">
          <Input placeholder="Update shop name" value={shopName} onChange={(e) => setShopName(e.target.value)} />
          <Button variant={submit ? "default" : "outline"} onClick={() => setSubmit((v) => !v)}>
            {submit ? "Will submit" : "Draft update"}
          </Button>
          <Button onClick={() => patch.mutate()} disabled={patch.isPending}>
            Save
          </Button>
        </div>
        <div className="mt-3">
          <Button variant="outline" onClick={() => createDoc.mutate()} disabled={createDoc.isPending}>
            Upload document (stub URL)
          </Button>
        </div>
      </SellerSectionCard>
      <SellerJsonView value={app.data} />
    </div>
  )
}

