import { BrandsClient } from "@/app/(storefront)/brands/brands-client"

export const metadata = {
  title: "Brands",
  description: "Browse brands in the public catalog.",
}

export default function Page({
  searchParams,
}: {
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <BrandsClient searchParams={searchParams ?? {}} />
}
