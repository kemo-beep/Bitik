import { SellerOrderDetailClient } from "@/app/seller/orders/[id]/seller-order-detail-client"

export const metadata = { title: "Order detail" }

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <SellerOrderDetailClient orderId={id} />
}
