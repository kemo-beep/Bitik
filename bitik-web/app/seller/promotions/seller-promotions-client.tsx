"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asString } from "@/lib/safe"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export function SellerPromotionsClient() {
  const qc = useQueryClient()
  const vouchers = useQuery({ queryKey: queryKeys.seller.vouchers(), queryFn: sellerApi.listVouchers })
  const [code, setCode] = useState("")
  const [discount, setDiscount] = useState("10")
  const refresh = () => qc.invalidateQueries({ queryKey: queryKeys.seller.vouchers() })
  const create = useMutation({
    mutationFn: () => sellerApi.createVoucher({ code: code.trim(), discount_percent: Number.parseInt(discount, 10) || 0 }),
    onSuccess: () => {
      setCode("")
      setDiscount("10")
      refresh()
    },
  })
  const remove = useMutation({ mutationFn: (id: string) => sellerApi.deleteVoucher(id), onSuccess: refresh })
  const items = asArray((vouchers.data as Record<string, unknown> | undefined)?.items ?? vouchers.data) ?? []

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Promotions</h1>
      <SellerSectionCard title="Create voucher">
        <div className="grid gap-2 md:grid-cols-3">
          <Input placeholder="Code" value={code} onChange={(e) => setCode(e.target.value)} />
          <Input placeholder="Discount %" value={discount} onChange={(e) => setDiscount(e.target.value)} />
          <Button onClick={() => create.mutate()} disabled={!code.trim() || create.isPending}>Create</Button>
        </div>
      </SellerSectionCard>
      <SellerSectionCard title="Voucher list">
        <div className="space-y-2">
          {items.map((v, i) => {
            const rec = v as Record<string, unknown>
            const id = asString(rec.id) ?? `${i}`
            return (
              <div key={id} className="flex items-center justify-between rounded border p-3">
                <span className="text-sm">{asString(rec.code) ?? id}</span>
                <Button size="sm" variant="destructive" onClick={() => remove.mutate(id)}>Delete</Button>
              </div>
            )
          })}
        </div>
        <div className="mt-3">
          <SellerJsonView value={vouchers.data} />
        </div>
      </SellerSectionCard>
    </div>
  )
}

