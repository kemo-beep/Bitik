"use client"

import * as React from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asString } from "@/lib/safe"
import {
  SellerJsonView,
  SellerSectionCard,
} from "@/components/seller/section-card"
import { Button } from "@/components/ui/button"
import { useRealtimeChannel } from "@/lib/realtime/use-realtime-channel"

export function SellerOrderDetailClient({ orderId }: { orderId: string }) {
  const qc = useQueryClient()
  const refresh = React.useCallback(() => {
    qc.invalidateQueries({ queryKey: queryKeys.seller.order(orderId) })
    qc.invalidateQueries({ queryKey: queryKeys.seller.orderShipments(orderId) })
  }, [orderId, qc])
  const order = useQuery({
    queryKey: queryKeys.seller.order(orderId),
    queryFn: () => sellerApi.getOrder(orderId),
  })
  const orderStatusChannel = useRealtimeChannel(
    "order-status",
    (event) => {
      if (String((event.data ?? {}).order_id ?? "") !== orderId) return
      refresh()
    },
    true
  )
  const items = useQuery({
    queryKey: queryKeys.seller.orderItems(orderId),
    queryFn: () => sellerApi.listOrderItems(orderId),
  })
  const shipments = useQuery({
    queryKey: queryKeys.seller.orderShipments(orderId),
    queryFn: () => sellerApi.listOrderShipments(orderId),
    refetchInterval: orderStatusChannel.connected ? false : 10_000,
  })
  const ship = useMutation({
    mutationFn: () => sellerApi.shipOrder(orderId, {}),
    onSuccess: refresh,
  })
  const accept = useMutation({
    mutationFn: () => sellerApi.acceptOrder(orderId),
    onSuccess: refresh,
  })
  const reject = useMutation({
    mutationFn: () => sellerApi.rejectOrder(orderId),
    onSuccess: refresh,
  })
  const pack = useMutation({
    mutationFn: () => sellerApi.packOrder(orderId),
    onSuccess: refresh,
  })
  const cancel = useMutation({
    mutationFn: () => sellerApi.cancelOrder(orderId),
    onSuccess: refresh,
  })
  const refund = useMutation({
    mutationFn: () => sellerApi.refundOrder(orderId),
    onSuccess: refresh,
  })
  const approveReturn = useMutation({
    mutationFn: () => sellerApi.approveReturn(orderId),
    onSuccess: refresh,
  })
  const rejectReturn = useMutation({
    mutationFn: () => sellerApi.rejectReturn(orderId),
    onSuccess: refresh,
  })
  const markReturnReceived = useMutation({
    mutationFn: () => sellerApi.markReturnReceived(orderId),
    onSuccess: refresh,
  })
  const markPacked = useMutation({
    mutationFn: (shipmentId: string) => sellerApi.markPacked(shipmentId),
    onSuccess: refresh,
  })
  const markShipped = useMutation({
    mutationFn: (shipmentId: string) =>
      sellerApi.markShipmentShipped(shipmentId),
    onSuccess: refresh,
  })
  const markDelivered = useMutation({
    mutationFn: (shipmentId: string) => sellerApi.markDelivered(shipmentId),
    onSuccess: refresh,
  })
  const label = useMutation({
    mutationFn: (shipmentId: string) =>
      sellerApi.createShipmentLabel(shipmentId, {}),
    onSuccess: refresh,
  })

  const shipmentItems =
    asArray(
      (shipments.data as Record<string, unknown> | undefined)?.items ??
        shipments.data
    ) ?? []

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Order detail</h1>
      <p className="text-xs text-muted-foreground">
        Realtime channel:{" "}
        {orderStatusChannel.connected ? "connected" : "polling fallback"}
      </p>
      <SellerSectionCard title="Order payload">
        <SellerJsonView value={order.data} />
      </SellerSectionCard>
      <SellerSectionCard title="Order items">
        <SellerJsonView
          value={
            asArray(
              (items.data as Record<string, unknown> | undefined)?.items ??
                items.data
            ) ?? []
          }
        />
      </SellerSectionCard>
      <SellerSectionCard title="Fulfillment actions">
        <div className="flex flex-wrap gap-2">
          <Button
            variant="outline"
            onClick={() => accept.mutate()}
            disabled={accept.isPending}
          >
            Accept
          </Button>
          <Button
            variant="outline"
            onClick={() => reject.mutate()}
            disabled={reject.isPending}
          >
            Reject
          </Button>
          <Button
            variant="outline"
            onClick={() => pack.mutate()}
            disabled={pack.isPending}
          >
            Pack
          </Button>
          <Button onClick={() => ship.mutate()} disabled={ship.isPending}>
            Ship
          </Button>
          <Button
            variant="outline"
            onClick={() => cancel.mutate()}
            disabled={cancel.isPending}
          >
            Cancel
          </Button>
          <Button
            variant="outline"
            onClick={() => refund.mutate()}
            disabled={refund.isPending}
          >
            Refund
          </Button>
          <Button
            variant="outline"
            onClick={() => approveReturn.mutate()}
            disabled={approveReturn.isPending}
          >
            Approve return
          </Button>
          <Button
            variant="outline"
            onClick={() => rejectReturn.mutate()}
            disabled={rejectReturn.isPending}
          >
            Reject return
          </Button>
          <Button
            variant="outline"
            onClick={() => markReturnReceived.mutate()}
            disabled={markReturnReceived.isPending}
          >
            Return received
          </Button>
        </div>
      </SellerSectionCard>
      <SellerSectionCard title="Shipments">
        <div className="space-y-2">
          {shipmentItems.map((s, i) => {
            const rec = s as Record<string, unknown>
            const id = asString(rec.id) ?? `${i}`
            return (
              <div key={id} className="rounded border p-3">
                <div className="flex flex-wrap gap-2">
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => markPacked.mutate(id)}
                  >
                    Pack
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => markShipped.mutate(id)}
                  >
                    Ship
                  </Button>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => markDelivered.mutate(id)}
                  >
                    Deliver
                  </Button>
                  <Button size="sm" onClick={() => label.mutate(id)}>
                    Print label
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
        <SellerJsonView value={shipments.data} />
      </SellerSectionCard>
    </div>
  )
}
