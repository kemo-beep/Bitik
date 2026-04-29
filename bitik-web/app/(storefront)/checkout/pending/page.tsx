import Link from "next/link"
import { routes } from "@/lib/routes"

export const dynamic = "force-dynamic"

export const metadata = { title: "Awaiting confirmation" }

export default async function Page({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>
}) {
  const sp = searchParams ? await searchParams : {}
  const orderId = typeof sp.order_id === "string" ? sp.order_id : ""
  return (
    <div className="mx-auto max-w-screen-md px-4 py-10">
      <h1 className="font-heading text-2xl font-semibold">Awaiting manual confirmation</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        Your payment is pending review. This usually takes a short while.
      </p>
      {orderId ? <p className="mt-2 text-sm">Order: <span className="font-mono">{orderId}</span></p> : null}
      <div className="mt-4 flex gap-2">
        <Link href={routes.account.orders} className="underline">View my orders</Link>
        {orderId ? <Link href={`${routes.storefront.checkoutPayment}?order_id=${encodeURIComponent(orderId)}`} className="underline">Back to payment</Link> : null}
      </div>
    </div>
  )
}
