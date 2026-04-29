import { CartClient } from "@/app/(storefront)/cart/cart-client"

export const metadata = { title: "Cart", description: "Manage cart items and proceed to checkout." }

export default function Page() {
  return <CartClient />
}
