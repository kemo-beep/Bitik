import { CategoryClient } from "@/app/(storefront)/categories/[id]/category-client"

export const metadata = {
  title: "Category",
  description: "Products in this category.",
}

export default function Page({
  params,
  searchParams,
}: {
  params: { id: string }
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <CategoryClient categoryId={params.id} searchParams={searchParams ?? {}} />
}
