"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export function SellerShippingClient() {
  const qc = useQueryClient()
  const [orderId, setOrderId] = useState("")
  const [shipmentId, setShipmentId] = useState("")
  const [provider, setProvider] = useState("manual")
  const [leadDays, setLeadDays] = useState("2")
  const settings = useQuery({ queryKey: ["seller", "shipping-settings"], queryFn: sellerApi.getShippingSettings })
  const saveSettings = useMutation({
    mutationFn: () =>
      sellerApi.patchShippingSettings({
        default_provider: provider.trim() || "manual",
        lead_time_days: Number.parseInt(leadDays, 10) || 2,
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["seller", "shipping-settings"] }),
  })
  const list = useMutation({ mutationFn: () => sellerApi.listOrderShipments(orderId.trim()) })
  const patch = useMutation({ mutationFn: () => sellerApi.patchShipment(shipmentId.trim(), { tracking_number: "TRK123" }) })
  const label = useMutation({ mutationFn: () => sellerApi.createShipmentLabel(shipmentId.trim(), {}) })
  const markShipped = useMutation({ mutationFn: () => sellerApi.markShipmentShipped(shipmentId.trim()) })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Shipping</h1>
      <SellerSectionCard title="Shipping settings">
        <div className="grid gap-2 md:grid-cols-3">
          <Input placeholder="Default provider" value={provider} onChange={(e) => setProvider(e.target.value)} />
          <Input placeholder="Lead time days" value={leadDays} onChange={(e) => setLeadDays(e.target.value)} />
          <Button onClick={() => saveSettings.mutate()} disabled={saveSettings.isPending}>Save settings</Button>
        </div>
        <div className="mt-3">
          <SellerJsonView value={settings.data} />
        </div>
      </SellerSectionCard>
      <SellerSectionCard title="Order shipments">
        <div className="flex gap-2">
          <Input placeholder="Order ID" value={orderId} onChange={(e) => setOrderId(e.target.value)} />
          <Button onClick={() => list.mutate()} disabled={!orderId.trim() || list.isPending}>Load shipments</Button>
        </div>
        {list.data ? <div className="mt-3"><SellerJsonView value={list.data} /></div> : null}
      </SellerSectionCard>
      <SellerSectionCard title="Shipment actions">
        <div className="flex gap-2">
          <Input placeholder="Shipment ID" value={shipmentId} onChange={(e) => setShipmentId(e.target.value)} />
          <Button variant="outline" onClick={() => patch.mutate()} disabled={!shipmentId.trim() || patch.isPending}>Edit</Button>
          <Button variant="outline" onClick={() => label.mutate()} disabled={!shipmentId.trim() || label.isPending}>Print label</Button>
          <Button onClick={() => markShipped.mutate()} disabled={!shipmentId.trim() || markShipped.isPending}>Mark shipped</Button>
        </div>
      </SellerSectionCard>
    </div>
  )
}

