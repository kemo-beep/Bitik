import { SellerProductEditorClient } from "@/app/seller/products/product-editor-client"

export const metadata = { title: "Edit product" }

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <SellerProductEditorClient productId={id} />
}
