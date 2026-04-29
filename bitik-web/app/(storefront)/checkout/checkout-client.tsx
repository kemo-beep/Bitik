"use client"

import { useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
  applyCheckoutVoucher,
  createCheckoutSession,
  getCheckoutSession,
  patchCheckoutAddress,
  patchCheckoutPaymentMethod,
  patchCheckoutShipping,
  placeOrderFromCheckoutSession,
  validateCheckoutSession,
} from "@/lib/api/buyer"
import { listAddresses } from "@/lib/api/account"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asRecord, asString } from "@/lib/safe"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { routes } from "@/lib/routes"
import { useAnalytics, analyticsEvents } from "@/lib/analytics"

function idem() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function CheckoutClient() {
  const router = useRouter()
  const qc = useQueryClient()
  const [checkoutSessionId, setCheckoutSessionId] = useState("")
  const [voucherCode, setVoucherCode] = useState("")
  const analytics = useAnalytics()

  const create = useMutation({
    mutationFn: () => createCheckoutSession({}),
    onSuccess: (data) => {
      const id = asString(data.checkout_session_id) ?? asString(data.id) ?? ""
      if (id) setCheckoutSessionId(id)
    },
  })

  useEffect(() => {
    create.mutate()
    analytics.track({ name: analyticsEvents.checkoutStart })
    // one-time session bootstrap
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const addresses = useQuery({ queryKey: ["buyer", "addresses"], queryFn: listAddresses })
  const session = useQuery({
    queryKey: queryKeys.buyer.checkoutSession(checkoutSessionId || "pending"),
    queryFn: () => getCheckoutSession(checkoutSessionId),
    enabled: checkoutSessionId !== "",
  })

  const refresh = () => {
    if (!checkoutSessionId) return
    qc.invalidateQueries({ queryKey: queryKeys.buyer.checkoutSession(checkoutSessionId) })
  }

  const saveAddress = useMutation({
    mutationFn: (addressId: string) => patchCheckoutAddress(checkoutSessionId, { shipping_address_id: addressId, billing_address_id: addressId }),
    onSuccess: refresh,
  })
  const saveShipping = useMutation({
    mutationFn: (shippingMethod: string) => patchCheckoutShipping(checkoutSessionId, { shipping_method: shippingMethod }),
    onSuccess: refresh,
  })
  const savePayment = useMutation({
    mutationFn: (paymentMethod: string) => {
      analytics.track({ name: analyticsEvents.paymentMethodSelected, properties: { payment_method: paymentMethod } })
      return patchCheckoutPaymentMethod(checkoutSessionId, { payment_method: paymentMethod })
    },
    onSuccess: refresh,
  })
  const applyVoucher = useMutation({
    mutationFn: () => applyCheckoutVoucher(checkoutSessionId, { code: voucherCode.trim() }),
    onSuccess: () => {
      setVoucherCode("")
      refresh()
    },
  })
  const validate = useMutation({
    mutationFn: () => validateCheckoutSession(checkoutSessionId, {}),
  })
  const place = useMutation({
    mutationFn: async () => placeOrderFromCheckoutSession(checkoutSessionId, {}, idem()),
    onSuccess: (data) => {
      analytics.track({
        name: analyticsEvents.placeOrder,
        properties: { checkout_session_id: checkoutSessionId, order_id: asString(data.order_id) ?? null },
      })
      const orderId = asString(data.order_id) ?? asString(data.id)
      const paymentStatus = asString(data.payment_status) ?? ""
      if (paymentStatus === "pending_manual") {
        router.push(`${routes.storefront.checkoutPending}?order_id=${encodeURIComponent(orderId ?? "")}`)
        return
      }
      if (orderId) router.push(routes.storefront.checkoutSuccess(orderId))
      else router.push(routes.storefront.checkoutPayment)
    },
  })

  const addressItems = addresses.data ?? []
  const selectedAddress = asString(asRecord(session.data)?.shipping_address_id) ?? ""
  const selectedShipping = asString(asRecord(session.data)?.shipping_method) ?? ""
  const selectedPayment = asString(asRecord(session.data)?.payment_method) ?? ""

  const canPlace = useMemo(
    () => checkoutSessionId !== "" && selectedAddress !== "" && selectedShipping !== "" && selectedPayment !== "",
    [checkoutSessionId, selectedAddress, selectedShipping, selectedPayment]
  )

  return (
    <div className="mx-auto max-w-screen-2xl px-4 py-8 lg:px-6">
      <h1 className="font-heading text-2xl font-semibold">Checkout</h1>
      <p className="mt-1 text-sm text-muted-foreground">Review details, validate, and place order.</p>

      {create.isPending ? <div className="mt-4 rounded border p-3 text-sm">Creating checkout session…</div> : null}

      <div className="mt-6 grid gap-4 lg:grid-cols-2">
        <section className="rounded-xl border p-4">
          <h2 className="font-medium">Address</h2>
          <div className="mt-3 space-y-2">
            {addressItems.map((a) => (
              <button
                key={a.id}
                className={`w-full rounded border p-3 text-left text-sm ${selectedAddress === a.id ? "border-primary" : ""}`}
                onClick={() => saveAddress.mutate(a.id)}
                type="button"
              >
                <div className="font-medium">{a.full_name}</div>
                <div className="text-muted-foreground">{a.address_line1}</div>
              </button>
            ))}
          </div>
        </section>

        <section className="rounded-xl border p-4">
          <h2 className="font-medium">Shipping</h2>
          <div className="mt-3 flex flex-wrap gap-2">
            {["standard", "express"].map((m) => (
              <Button key={m} variant={selectedShipping === m ? "default" : "outline"} onClick={() => saveShipping.mutate(m)}>
                {m}
              </Button>
            ))}
          </div>
        </section>

        <section className="rounded-xl border p-4">
          <h2 className="font-medium">Payment method</h2>
          <div className="mt-3 flex flex-wrap gap-2">
            {["wave_manual", "pod"].map((m) => (
              <Button key={m} variant={selectedPayment === m ? "default" : "outline"} onClick={() => savePayment.mutate(m)}>
                {m === "wave_manual" ? "Wave manual" : "POD"}
              </Button>
            ))}
          </div>
          <p className="mt-2 text-xs text-muted-foreground">POD availability depends on backend eligibility checks.</p>
        </section>

        <section className="rounded-xl border p-4">
          <h2 className="font-medium">Voucher</h2>
          <div className="mt-3 flex gap-2">
            <Input value={voucherCode} onChange={(e) => setVoucherCode(e.target.value)} placeholder="Voucher code" />
            <Button onClick={() => applyVoucher.mutate()} disabled={!voucherCode.trim() || applyVoucher.isPending}>
              Apply
            </Button>
          </div>
        </section>
      </div>

      <section className="mt-4 rounded-xl border p-4">
        <h2 className="font-medium">Order summary</h2>
        <pre className="mt-2 max-h-56 overflow-auto rounded bg-muted p-2 text-xs">
          {JSON.stringify(asRecord(session.data)?.summary ?? asRecord(session.data) ?? {}, null, 2)}
        </pre>
        <div className="mt-4 flex flex-wrap gap-2">
          <Button variant="outline" onClick={() => validate.mutate()} disabled={!checkoutSessionId || validate.isPending}>
            Validate
          </Button>
          <Button onClick={() => place.mutate()} disabled={!canPlace || place.isPending}>
            Place order
          </Button>
        </div>
        {asArray(asRecord(validate.data)?.errors)?.length ? (
          <div className="mt-2 text-sm text-destructive">Validation returned errors. Please review selections.</div>
        ) : null}
      </section>
    </div>
  )
}

