"use client"

import * as React from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  disputeBuyerOrder,
  getBuyerOrder,
  getBuyerOrderInvoice,
  getBuyerOrderTracking,
  listBuyerOrderItems,
  listBuyerOrderStatusHistory,
  requestBuyerOrderRefund,
  requestBuyerOrderReturn,
} from "@/lib/api/buyer"
import { queryKeys } from "@/lib/queryKeys"
import { asArray } from "@/lib/safe"
import { Button } from "@/components/ui/button"
import { useRealtimeChannel } from "@/lib/realtime/use-realtime-channel"

export function OrderDetailClient({ orderId }: { orderId: string }) {
  const qc = useQueryClient()
  const refresh = React.useCallback(() => {
    qc.invalidateQueries({ queryKey: queryKeys.buyer.order(orderId) })
    qc.invalidateQueries({
      queryKey: queryKeys.buyer.orderStatusHistory(orderId),
    })
  }, [orderId, qc])
  const order = useQuery({
    queryKey: queryKeys.buyer.order(orderId),
    queryFn: () => getBuyerOrder(orderId),
  })
  const orderStatusChannel = useRealtimeChannel(
    "order-status",
    (event) => {
      if (String((event.data ?? {}).order_id ?? "") !== orderId) return
      refresh()
      qc.invalidateQueries({ queryKey: queryKeys.buyer.orderTracking(orderId) })
    },
    true
  )
  const items = useQuery({
    queryKey: queryKeys.buyer.orderItems(orderId),
    queryFn: () => listBuyerOrderItems(orderId),
  })
  const history = useQuery({
    queryKey: queryKeys.buyer.orderStatusHistory(orderId),
    queryFn: () => listBuyerOrderStatusHistory(orderId),
    refetchInterval: orderStatusChannel.connected ? false : 10_000,
  })
  const tracking = useQuery({
    queryKey: queryKeys.buyer.orderTracking(orderId),
    queryFn: () => getBuyerOrderTracking(orderId),
    refetchInterval: orderStatusChannel.connected ? false : 10_000,
  })
  const invoice = useQuery({
    queryKey: queryKeys.buyer.orderInvoice(orderId),
    queryFn: () => getBuyerOrderInvoice(orderId),
  })

  const refund = useMutation({
    mutationFn: () =>
      requestBuyerOrderRefund(orderId, { reason: "Requested from web" }),
    onSuccess: refresh,
  })
  const returns = useMutation({
    mutationFn: () =>
      requestBuyerOrderReturn(orderId, { reason: "Requested from web" }),
    onSuccess: refresh,
  })
  const dispute = useMutation({
    mutationFn: () =>
      disputeBuyerOrder(orderId, { reason: "Dispute opened from web" }),
    onSuccess: refresh,
  })

  return (
    <div className="mx-auto max-w-screen-xl px-4 py-8">
      <h1 className="font-heading text-2xl font-semibold">Order detail</h1>
      <p className="mt-1 text-xs text-muted-foreground">
        Realtime channel:{" "}
        {orderStatusChannel.connected ? "connected" : "polling fallback"}
      </p>
      <pre className="mt-4 max-h-56 overflow-auto rounded bg-muted p-3 text-xs">
        {JSON.stringify(order.data ?? {}, null, 2)}
      </pre>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button
          variant="outline"
          onClick={() => refund.mutate()}
          disabled={refund.isPending}
        >
          Request refund
        </Button>
        <Button
          variant="outline"
          onClick={() => returns.mutate()}
          disabled={returns.isPending}
        >
          Request return
        </Button>
        <Button
          variant="destructive"
          onClick={() => dispute.mutate()}
          disabled={dispute.isPending}
        >
          Open dispute
        </Button>
      </div>

      <section className="mt-6 rounded-xl border p-4">
        <h2 className="font-medium">Items</h2>
        <pre className="mt-2 max-h-56 overflow-auto rounded bg-muted p-3 text-xs">
          {JSON.stringify(
            asArray(
              (items.data as Record<string, unknown> | undefined)?.items ??
                items.data
            ) ?? [],
            null,
            2
          )}
        </pre>
      </section>

      <section className="mt-4 rounded-xl border p-4">
        <h2 className="font-medium">Status history</h2>
        <pre className="mt-2 max-h-56 overflow-auto rounded bg-muted p-3 text-xs">
          {JSON.stringify(
            asArray(
              (history.data as Record<string, unknown> | undefined)?.items ??
                history.data
            ) ?? [],
            null,
            2
          )}
        </pre>
      </section>

      <section className="mt-4 grid gap-4 md:grid-cols-2">
        <div className="rounded-xl border p-4">
          <h2 className="font-medium">Tracking</h2>
          <pre className="mt-2 max-h-56 overflow-auto rounded bg-muted p-3 text-xs">
            {JSON.stringify(tracking.data ?? {}, null, 2)}
          </pre>
        </div>
        <div className="rounded-xl border p-4">
          <h2 className="font-medium">Invoice</h2>
          <pre className="mt-2 max-h-56 overflow-auto rounded bg-muted p-3 text-xs">
            {JSON.stringify(invoice.data ?? {}, null, 2)}
          </pre>
        </div>
      </section>
    </div>
  )
}
