"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asString } from "@/lib/safe"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export function SellerWalletClient() {
  const qc = useQueryClient()
  const wallet = useQuery({ queryKey: queryKeys.seller.wallet(), queryFn: sellerApi.getWallet })
  const txns = useQuery({ queryKey: queryKeys.seller.walletTransactions(), queryFn: () => sellerApi.listWalletTransactions() })
  const payouts = useQuery({ queryKey: queryKeys.seller.payouts(), queryFn: () => sellerApi.listPayouts() })
  const bankAccounts = useQuery({ queryKey: ["seller", "bank-accounts"], queryFn: sellerApi.listBankAccounts })
  const [amount, setAmount] = useState("0")
  const [bankName, setBankName] = useState("")
  const [accountName, setAccountName] = useState("")
  const [accountNumber, setAccountNumber] = useState("")
  const requestPayout = useMutation({
    mutationFn: () => sellerApi.requestPayout({ amount_cents: Number.parseInt(amount, 10) || 0 }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.seller.payouts() })
      qc.invalidateQueries({ queryKey: queryKeys.seller.wallet() })
    },
  })
  const refreshBank = () => qc.invalidateQueries({ queryKey: ["seller", "bank-accounts"] })
  const addBank = useMutation({
    mutationFn: () =>
      sellerApi.createBankAccount({
        bank_name: bankName.trim(),
        account_name: accountName.trim(),
        account_number: accountNumber.trim(),
        is_default: true,
      }),
    onSuccess: () => {
      setBankName("")
      setAccountName("")
      setAccountNumber("")
      refreshBank()
    },
  })
  const setDefault = useMutation({ mutationFn: (id: string) => sellerApi.setDefaultBankAccount(id), onSuccess: refreshBank })
  const removeBank = useMutation({ mutationFn: (id: string) => sellerApi.deleteBankAccount(id), onSuccess: refreshBank })
  const bankItems = asArray((bankAccounts.data as Record<string, unknown> | undefined)?.items ?? bankAccounts.data) ?? []

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">Wallet and payouts</h1>
      <SellerSectionCard title="Wallet summary">
        <SellerJsonView value={wallet.data} />
      </SellerSectionCard>
      <SellerSectionCard title="Transactions">
        <SellerJsonView value={txns.data} />
      </SellerSectionCard>
      <SellerSectionCard title="Payout request">
        <div className="flex gap-2">
          <Input value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="Amount (cents)" />
          <Button onClick={() => requestPayout.mutate()} disabled={requestPayout.isPending}>Request payout</Button>
        </div>
        <div className="mt-3">
          <SellerJsonView value={payouts.data} />
        </div>
      </SellerSectionCard>
      <SellerSectionCard title="Bank accounts">
        <div className="grid gap-2 md:grid-cols-4">
          <Input placeholder="Bank name" value={bankName} onChange={(e) => setBankName(e.target.value)} />
          <Input placeholder="Account name" value={accountName} onChange={(e) => setAccountName(e.target.value)} />
          <Input placeholder="Account number" value={accountNumber} onChange={(e) => setAccountNumber(e.target.value)} />
          <Button onClick={() => addBank.mutate()} disabled={!bankName.trim() || !accountName.trim() || !accountNumber.trim() || addBank.isPending}>
            Add account
          </Button>
        </div>
        <div className="mt-3 space-y-2">
          {bankItems.map((it, idx) => {
            const rec = it as Record<string, unknown>
            const id = asString(rec.id) ?? `${idx}`
            return (
              <div key={id} className="flex items-center justify-between rounded border p-3">
                <span className="text-sm">{asString(rec.bank_name) ?? "Bank"} · {asString(rec.account_number_masked) ?? "****"}</span>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => setDefault.mutate(id)}>Set default</Button>
                  <Button size="sm" variant="destructive" onClick={() => removeBank.mutate(id)}>Delete</Button>
                </div>
              </div>
            )
          })}
        </div>
      </SellerSectionCard>
    </div>
  )
}

