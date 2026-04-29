import { SearchClient } from "@/app/(storefront)/search/search-client"

export const metadata = {
  title: "Search",
  description: "Search products with filters and sorting.",
}

export default function Page({
  searchParams,
}: {
  searchParams?: Record<string, string | string[] | undefined>
}) {
  return <SearchClient searchParams={searchParams ?? {}} />
}
