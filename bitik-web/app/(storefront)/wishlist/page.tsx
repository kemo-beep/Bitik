import { WishlistClient } from "@/app/(storefront)/wishlist/wishlist-client"

export const metadata = { title: "Wishlist", description: "Saved products and quick move-to-cart." }

export default function Page() {
  return <WishlistClient />
}
