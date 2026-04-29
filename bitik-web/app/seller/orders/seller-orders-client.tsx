"use client"

import Link from "next/link"
import { useQuery } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asString } from "@/lib/safe"
import { SellerSectionCard } from "@/components/seller/section-card"
import { routes } from "@/lib/routes"

export function SellerOrdersClient() {
  const orders = useQuery({ queryKey: queryKeys.seller.orders(), queryFn: () => sellerApi.listOrders() })
  const items = asArray((orders.data as Record<string, unknown> | undefined)?.items ?? orders.data) ?? []

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Orders</h1>
      <SellerSectionCard title="Seller orders">
        <div className="space-y-2">
          {orders.isLoading ? <div className="text-sm">Loading orders…</div> : null}
          {items.map((o, i) => {
            const rec = o as Record<string, unknown>
            const id = asString(rec.id) ?? `${i}`
            const status = asString(rec.status) ?? "unknown"
            return (
              <div key={id} className="rounded border p-3">
                <div className="flex items-center justify-between gap-2">
                  <Link className="font-medium hover:underline" href={routes.seller.order(id)}>
                    Order {id}
                  </Link>
                  <span className="text-sm text-muted-foreground">{status}</span>
                </div>
              </div>
            )
          })}
        </div>
      </SellerSectionCard>
    </div>
  )
}

