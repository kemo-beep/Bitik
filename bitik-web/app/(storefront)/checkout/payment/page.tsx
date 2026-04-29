import { PaymentClient } from "@/app/(storefront)/checkout/payment/payment-client"

export const metadata = { title: "Payment", description: "Wave manual instructions and payment edge-case handling." }

export default function Page({
  searchParams,
}: {
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <PaymentClient searchParams={searchParams ?? {}} />
}
