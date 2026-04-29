import { SellerProductEditorClient } from "@/app/seller/products/product-editor-client"

export const metadata = { title: "Edit product" }

export default function Page({ params }: { params: { id: string } }) {
  return <SellerProductEditorClient productId={params.id} />
}
