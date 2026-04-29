import { BrandClient } from "@/app/(storefront)/brands/[id]/brand-client"

export const metadata = {
  title: "Brand",
  description: "Products from this brand.",
}

export default async function Page({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>
  searchParams?: Promise<Record<string, string | string[] | undefined>>
}) {
  const { id } = await params
  const resolvedSearchParams = searchParams ? await searchParams : {}
  return <BrandClient brandId={id} searchParams={resolvedSearchParams} />
}
