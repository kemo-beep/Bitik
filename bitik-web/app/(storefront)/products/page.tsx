import { ProductsClient } from "@/app/(storefront)/products/products-client"

export const metadata = {
  title: "Products",
  description: "Browse products with filters and sorting.",
}

export default function Page({
  searchParams,
}: {
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <ProductsClient searchParams={searchParams ?? {}} />
}
