"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asString } from "@/lib/safe"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export function SellerInventoryClient() {
  const qc = useQueryClient()
  const inventory = useQuery({ queryKey: queryKeys.seller.inventory(), queryFn: () => sellerApi.listInventory() })
  const low = useQuery({ queryKey: queryKeys.seller.inventoryLowStock(), queryFn: sellerApi.listLowStock })
  const [delta, setDelta] = useState("1")
  const refresh = () => {
    qc.invalidateQueries({ queryKey: queryKeys.seller.inventory() })
    qc.invalidateQueries({ queryKey: queryKeys.seller.inventoryLowStock() })
  }

  const adjust = useMutation({
    mutationFn: ({ id }: { id: string }) =>
      sellerApi.adjustInventory(id, {
        delta: Number.parseInt(delta, 10) || 0,
        reason: "manual_adjustment",
      }),
    onSuccess: refresh,
  })
  const bulk = useMutation({
    mutationFn: () =>
      sellerApi.bulkUpdateInventory({
        updates: [
          { inventory_id: "00000000-0000-0000-0000-000000000000", quantity: 10 },
        ],
      }),
  })

  const items = asArray((inventory.data as Record<string, unknown> | undefined)?.items ?? inventory.data) ?? []

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Inventory</h1>
      <SellerSectionCard title="Inventory list">
        <div className="mb-3 flex items-center gap-2">
          <Input className="w-32" value={delta} onChange={(e) => setDelta(e.target.value)} />
          <Button variant="outline" onClick={() => bulk.mutate()} disabled={bulk.isPending}>Bulk update (sample payload)</Button>
        </div>
        <div className="space-y-2">
          {items.map((it, i) => {
            const rec = it as Record<string, unknown>
            const id = asString(rec.id) ?? `${i}`
            return (
              <div key={id} className="flex items-center justify-between rounded border p-3">
                <span className="text-sm font-medium">{id}</span>
                <Button size="sm" onClick={() => adjust.mutate({ id })}>Adjust</Button>
              </div>
            )
          })}
        </div>
      </SellerSectionCard>
      <SellerSectionCard title="Low stock alerts">
        <SellerJsonView value={low.data} />
      </SellerSectionCard>
    </div>
  )
}

