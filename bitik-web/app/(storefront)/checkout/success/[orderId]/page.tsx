export const metadata = { title: "Order placed" }

import Link from "next/link"
import { routes } from "@/lib/routes"

export default async function Page({ params }: { params: Promise<{ orderId: string }> }) {
  const { orderId } = await params
  return (
    <div className="mx-auto max-w-screen-md px-4 py-10">
      <h1 className="font-heading text-2xl font-semibold">Order placed</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        Your order has been created successfully.
      </p>
      <p className="mt-2 text-sm">
        Order ID: <span className="font-mono">{orderId}</span>
      </p>
      <div className="mt-4 flex gap-2">
        <Link href={routes.account.order(orderId)} className="underline">View order detail</Link>
        <Link href={routes.storefront.products} className="underline">Continue shopping</Link>
      </div>
    </div>
  )
}
