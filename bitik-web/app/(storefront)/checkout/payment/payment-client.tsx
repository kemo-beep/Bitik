"use client"

import { useMemo, useState } from "react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import { useMutation, useQuery } from "@tanstack/react-query"
import { cancelBuyerPayment, confirmBuyerPayment, createBuyerPaymentIntent, getBuyerPayment } from "@/lib/api/buyer"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { asString } from "@/lib/safe"
import { routes } from "@/lib/routes"

function idem() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function PaymentClient({
  searchParams,
}: {
  searchParams: Record<string, string | string[] | undefined>
}) {
  const live = useSearchParams()
  const orderId = live.get("order_id") ?? (typeof searchParams.order_id === "string" ? searchParams.order_id : "")
  const paymentId = live.get("payment_id") ?? (typeof searchParams.payment_id === "string" ? searchParams.payment_id : "")
  const [reference, setReference] = useState("")
  const [amount, setAmount] = useState("")

  const createIntent = useMutation({
    mutationFn: () => createBuyerPaymentIntent({ order_id: orderId, channel: "wave_manual" }, idem()),
  })
  const confirm = useMutation({
    mutationFn: (id: string) =>
      confirmBuyerPayment({ payment_id: id, wave_reference: reference.trim(), amount: amount.trim() }),
  })
  const cancel = useMutation({
    mutationFn: (id: string) => cancelBuyerPayment(id, {}),
  })

  const effectivePaymentId = asString(createIntent.data?.payment_id) ?? paymentId
  const payment = useQuery({
    queryKey: ["buyer", "payments", effectivePaymentId || "none"],
    queryFn: () => getBuyerPayment(effectivePaymentId),
    enabled: effectivePaymentId !== "",
  })

  const paymentStatus = asString((payment.data as Record<string, unknown> | undefined)?.status) ?? ""
  const needsConfirm = useMemo(() => paymentStatus === "" || paymentStatus === "pending" || paymentStatus === "awaiting_evidence", [paymentStatus])

  return (
    <div className="mx-auto max-w-screen-md px-4 py-8">
      <h1 className="font-heading text-2xl font-semibold">Payment</h1>
      <p className="mt-1 text-sm text-muted-foreground">Wave manual payment instructions and retry/cancel states.</p>

      <div className="mt-6 rounded-xl border p-4">
        <div className="font-medium">Wave manual instructions</div>
        <ul className="mt-2 list-disc pl-5 text-sm text-muted-foreground">
          <li>Transfer the exact order amount to Wave merchant account.</li>
          <li>Use the order reference in transfer note.</li>
          <li>Submit reference + amount below for manual confirmation.</li>
        </ul>
        <div className="mt-3 text-sm">
          <div>Order reference: <span className="font-mono">{orderId || "N/A"}</span></div>
        </div>
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        <Button onClick={() => createIntent.mutate()} disabled={!orderId || createIntent.isPending}>
          Create payment intent
        </Button>
        {effectivePaymentId ? (
          <Button variant="outline" onClick={() => cancel.mutate(effectivePaymentId)} disabled={cancel.isPending}>
            Cancel payment
          </Button>
        ) : null}
      </div>

      {needsConfirm && effectivePaymentId ? (
        <div className="mt-4 rounded-xl border p-4">
          <h2 className="font-medium">Confirm payment evidence</h2>
          <div className="mt-3 grid gap-2 sm:grid-cols-2">
            <Input placeholder="Wave transaction reference" value={reference} onChange={(e) => setReference(e.target.value)} />
            <Input placeholder="Amount paid" value={amount} onChange={(e) => setAmount(e.target.value)} />
          </div>
          <Button
            className="mt-3"
            onClick={() => confirm.mutate(effectivePaymentId)}
            disabled={!reference.trim() || !amount.trim() || confirm.isPending}
          >
            Submit confirmation
          </Button>
        </div>
      ) : null}

      {paymentStatus === "pending_manual" || paymentStatus === "pending" ? (
        <div className="mt-4 rounded-xl border bg-amber-50 p-4 text-sm">
          Payment is pending manual confirmation.{" "}
          <Link className="underline" href={`${routes.storefront.checkoutPending}?order_id=${encodeURIComponent(orderId)}`}>
            Go to pending page
          </Link>
        </div>
      ) : null}

      {paymentStatus === "rejected" || paymentStatus === "failed" || paymentStatus === "timeout" ? (
        <div className="mt-4 rounded-xl border bg-destructive/10 p-4 text-sm text-destructive">
          Payment {paymentStatus}. You can retry by creating a new intent or cancel this payment.
        </div>
      ) : null}
    </div>
  )
}

