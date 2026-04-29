"use client"

import Link from "next/link"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { cancelBuyerOrder, confirmBuyerOrderReceived, listBuyerOrders } from "@/lib/api/buyer"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asString } from "@/lib/safe"
import { Button } from "@/components/ui/button"
import { routes } from "@/lib/routes"

export function OrdersClient({
  searchParams,
}: {
  searchParams: Record<string, string | string[] | undefined>
}) {
  const status = typeof searchParams.status === "string" ? searchParams.status : ""
  const qc = useQueryClient()
  const orders = useQuery({
    queryKey: queryKeys.buyer.orders({ status }),
    queryFn: () => listBuyerOrders({ status: status || undefined }),
  })
  const refresh = () => qc.invalidateQueries({ queryKey: queryKeys.buyer.orders({ status }) })
  const cancel = useMutation({
    mutationFn: (id: string) => cancelBuyerOrder(id, { reason: "User cancelled from web" }),
    onSuccess: refresh,
  })
  const confirm = useMutation({
    mutationFn: (id: string) => confirmBuyerOrderReceived(id, {}),
    onSuccess: refresh,
  })

  const items = asArray((orders.data as Record<string, unknown> | undefined)?.items ?? orders.data) ?? []

  return (
    <div className="mx-auto max-w-screen-xl px-4 py-8">
      <h1 className="font-heading text-2xl font-semibold">My orders</h1>
      <div className="mt-4 flex flex-wrap gap-2">
        {["", "pending", "paid", "shipped", "delivered", "cancelled"].map((s) => (
          <Link
            key={s || "all"}
            href={s ? `${routes.account.orders}?status=${encodeURIComponent(s)}` : routes.account.orders}
            className={`rounded border px-3 py-1 text-sm ${status === s ? "border-primary" : ""}`}
          >
            {s || "all"}
          </Link>
        ))}
      </div>

      <div className="mt-6 space-y-3">
        {orders.isLoading ? <div className="rounded border p-4 text-sm">Loading orders…</div> : null}
        {!orders.isLoading && items.length === 0 ? (
          <div className="rounded border p-6 text-sm text-muted-foreground">No orders found.</div>
        ) : null}
        {items.map((o, i) => {
          const rec = (o as Record<string, unknown>) ?? {}
          const id = asString(rec.id) ?? `${i}`
          const st = asString(rec.status) ?? "unknown"
          return (
            <div key={id} className="rounded-xl border p-4">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <div className="font-medium">Order {id}</div>
                  <div className="text-sm text-muted-foreground">Status: {st}</div>
                </div>
                <div className="flex gap-2">
                  <Button variant="outline" nativeButton={false} render={<Link href={routes.account.order(id)}>Detail</Link>} />
                  <Button variant="outline" onClick={() => confirm.mutate(id)} disabled={confirm.isPending}>Confirm received</Button>
                  <Button variant="destructive" onClick={() => cancel.mutate(id)} disabled={cancel.isPending}>Cancel</Button>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

