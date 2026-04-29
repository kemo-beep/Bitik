import { OrderDetailClient } from "@/app/account/orders/[id]/order-detail-client"

export const metadata = { title: "Order detail" }

export default function Page({ params }: { params: { id: string } }) {
  return <OrderDetailClient orderId={params.id} />
}
