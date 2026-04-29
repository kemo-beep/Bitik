import { CheckoutClient } from "@/app/(storefront)/checkout/checkout-client"

export const metadata = { title: "Checkout", description: "Address, shipping, payment, validation and place order." }

export default function Page() {
  return <CheckoutClient />
}
