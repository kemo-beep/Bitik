import { CategoryClient } from "@/app/(storefront)/categories/[id]/category-client"

export const metadata = {
  title: "Category",
  description: "Products in this category.",
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
  return <CategoryClient categoryId={id} searchParams={resolvedSearchParams} />
}
