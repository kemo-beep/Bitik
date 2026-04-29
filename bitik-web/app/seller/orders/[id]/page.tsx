import { SellerOrderDetailClient } from "@/app/seller/orders/[id]/seller-order-detail-client"

export const metadata = { title: "Order detail" }

export default function Page({ params }: { params: { id: string } }) {
  return <SellerOrderDetailClient orderId={params.id} />
}
