import { OrderDetailClient } from "@/app/account/orders/[id]/order-detail-client"

export const metadata = { title: "Order detail" }

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <OrderDetailClient orderId={id} />
}
