import { BrandClient } from "@/app/(storefront)/brands/[id]/brand-client"

export const metadata = {
  title: "Brand",
  description: "Products from this brand.",
}

export default function Page({
  params,
  searchParams,
}: {
  params: { id: string }
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <BrandClient brandId={params.id} searchParams={searchParams ?? {}} />
}
