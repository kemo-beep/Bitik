import { OrdersClient } from "@/app/account/orders/orders-client"

export const metadata = { title: "Orders" }

export default function Page({
  searchParams,
}: {
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <OrdersClient searchParams={searchParams ?? {}} />
}
