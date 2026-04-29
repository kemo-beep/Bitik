"use client"

import { useState } from "react"
import { useMutation } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { SellerSectionCard, SellerJsonView } from "@/components/seller/section-card"
import { Button } from "@/components/ui/button"

export function SellerSettingsClient() {
  const [autoAccept, setAutoAccept] = useState(false)
  const [enablePOD, setEnablePOD] = useState(true)
  const save = useMutation({
    mutationFn: () => sellerApi.patchSettings({ auto_accept_orders: autoAccept, enable_pod: enablePOD }),
  })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Settings</h1>
      <SellerSectionCard title="Seller settings">
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={autoAccept} onChange={(e) => setAutoAccept(e.target.checked)} />
          Auto-accept new orders
        </label>
        <label className="mt-2 flex items-center gap-2 text-sm">
          <input type="checkbox" checked={enablePOD} onChange={(e) => setEnablePOD(e.target.checked)} />
          Enable POD payments
        </label>
        <Button className="mt-3" onClick={() => save.mutate()} disabled={save.isPending}>
          Save settings
        </Button>
      </SellerSectionCard>
      {save.data ? <SellerJsonView value={save.data} /> : null}
    </div>
  )
}

