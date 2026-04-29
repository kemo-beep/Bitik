"use client"

import { useState } from "react"
import Link from "next/link"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { applyCartVoucher, clearCart, deleteCartItem, getCart, updateCartItem } from "@/lib/api/buyer"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asNumber, asRecord, asString } from "@/lib/safe"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { routes } from "@/lib/routes"
import { formatMoney } from "@/lib/format"

export function CartClient() {
  const qc = useQueryClient()
  const [voucherCode, setVoucherCode] = useState("")
  const cart = useQuery({
    queryKey: queryKeys.buyer.cart(),
    queryFn: getCart,
    staleTime: 20_000,
    gcTime: 10 * 60_000,
  })

  const items = asArray(asRecord(cart.data)?.items) ?? []
  const totals = asRecord(asRecord(cart.data)?.totals) ?? {}
  const currency = asString(totals.currency) ?? "MMK"
  const total = asNumber(totals.total_amount) ?? 0
  const selectedCount = asNumber(asRecord(cart.data)?.selected_count) ?? items.length

  const outWarnings: string[] = []
  for (const item of items) {
    const rec = asRecord(item) ?? {}
    if (asRecord(rec.inventory_warning)) outWarnings.push("Some items have limited inventory.")
    if (asRecord(rec.price_change_warning)) outWarnings.push("Some item prices changed.")
  }
  const warnings = Array.from(new Set(outWarnings))

  const refresh = () => qc.invalidateQueries({ queryKey: queryKeys.buyer.cart() })

  const clear = useMutation({
    mutationFn: clearCart,
    onSuccess: refresh,
  })
  const patchItem = useMutation({
    mutationFn: ({ id, qty }: { id: string; qty: number }) => updateCartItem(id, { quantity: qty }),
    onSuccess: refresh,
  })
  const removeItem = useMutation({
    mutationFn: (id: string) => deleteCartItem(id),
    onSuccess: refresh,
  })
  const applyVoucher = useMutation({
    mutationFn: () => applyCartVoucher({ code: voucherCode.trim() }),
    onSuccess: () => {
      setVoucherCode("")
      refresh()
    },
  })

  return (
    <div className="mx-auto max-w-screen-2xl px-4 py-8 lg:px-6">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="font-heading text-2xl font-semibold">Cart</h1>
          <p className="text-sm text-muted-foreground">{selectedCount} selected items</p>
        </div>
        <Button variant="outline" onClick={() => clear.mutate()} disabled={clear.isPending || !items.length}>
          Clear cart
        </Button>
      </div>

      {warnings.length ? (
        <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
          {warnings.join(" ")}
        </div>
      ) : null}

      <div className="mt-6 grid gap-4 lg:grid-cols-[1fr_320px]">
        <div className="space-y-3">
          {cart.isLoading ? <div className="rounded border p-4 text-sm">Loading cart…</div> : null}
          {!cart.isLoading && items.length === 0 ? (
            <div className="rounded border p-6 text-sm text-muted-foreground">Your cart is empty.</div>
          ) : null}
          {items.map((item, i) => {
            const rec = asRecord(item) ?? {}
            const id = asString(rec.id) ?? `${i}`
            const name = asString(rec.product_name) ?? asString(rec.name) ?? "Item"
            const qty = asNumber(rec.quantity) ?? 1
            const subtotal = asNumber(rec.subtotal_amount) ?? 0
            return (
              <div key={id} className="rounded-xl border p-4">
                <div className="flex items-center justify-between gap-4">
                  <div>
                    <div className="font-medium">{name}</div>
                    <div className="text-sm text-muted-foreground">{formatMoney(subtotal, { currency })}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Input
                      type="number"
                      min={1}
                      defaultValue={String(qty)}
                      className="w-20"
                      onBlur={(e) => {
                        const next = Number.parseInt(e.target.value, 10)
                        if (!Number.isFinite(next) || next < 1 || next === qty) return
                        patchItem.mutate({ id, qty: next })
                      }}
                    />
                    <Button variant="outline" onClick={() => removeItem.mutate(id)} disabled={removeItem.isPending}>
                      Remove
                    </Button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        <aside className="rounded-xl border p-4">
          <div className="text-sm text-muted-foreground">Cart total</div>
          <div className="mt-1 text-2xl font-semibold">{formatMoney(total, { currency })}</div>

          <div className="mt-4 flex gap-2">
            <Input
              placeholder="Voucher code"
              value={voucherCode}
              onChange={(e) => setVoucherCode(e.target.value)}
            />
            <Button onClick={() => applyVoucher.mutate()} disabled={!voucherCode.trim() || applyVoucher.isPending}>
              Apply
            </Button>
          </div>

          <Button
            className="mt-4 w-full"
            nativeButton={false}
            render={<Link href={routes.storefront.checkout}>Proceed to checkout</Link>}
            disabled={items.length === 0}
          />

          <p className="mt-3 text-xs text-muted-foreground">Guest cart merge is supported after login via backend API.</p>
        </aside>
      </div>
    </div>
  )
}

