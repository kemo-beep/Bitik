import { CategoriesClient } from "@/app/(storefront)/categories/categories-client"

export const metadata = {
  title: "Categories",
  description: "Browse all categories.",
}

export default function Page() {
  return <CategoriesClient />
}
