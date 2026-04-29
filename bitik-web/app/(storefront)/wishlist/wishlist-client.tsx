"use client"

import Link from "next/link"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { addCartItem, listWishlistProductIds, removeWishlistProductId } from "@/lib/api/buyer"
import { getProduct } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { asString } from "@/lib/safe"
import { Button } from "@/components/ui/button"
import { routes } from "@/lib/routes"

export function WishlistClient() {
  const qc = useQueryClient()
  const wishlist = useQuery({
    queryKey: queryKeys.buyer.wishlist(),
    queryFn: async () => {
      const ids = listWishlistProductIds()
      const products = await Promise.all(ids.map((id) => getProduct(id).catch(() => null)))
      return products.filter(Boolean)
    },
  })

  const refresh = () => qc.invalidateQueries({ queryKey: queryKeys.buyer.wishlist() })
  const moveToCart = useMutation({
    mutationFn: async (productId: string) => {
      await addCartItem({ product_id: productId, quantity: 1 })
      removeWishlistProductId(productId)
    },
    onSuccess: refresh,
  })

  return (
    <div className="mx-auto max-w-screen-2xl px-4 py-8 lg:px-6">
      <h1 className="font-heading text-2xl font-semibold">Wishlist</h1>
      <p className="mt-1 text-sm text-muted-foreground">Saved products. Move them to cart when ready.</p>

      <div className="mt-6 space-y-3">
        {wishlist.isLoading ? <div className="rounded border p-4 text-sm">Loading wishlist…</div> : null}
        {!wishlist.isLoading && (wishlist.data?.length ?? 0) === 0 ? (
          <div className="rounded border p-6 text-sm text-muted-foreground">Wishlist is empty.</div>
        ) : null}
        {(wishlist.data ?? []).map((p, i) => {
          const product = p as Record<string, unknown>
          const id = asString(product.id) ?? `${i}`
          const name = asString(product.name) ?? "Product"
          return (
            <div key={id} className="flex items-center justify-between rounded-xl border p-4">
              <Link href={routes.storefront.product(id)} className="font-medium hover:underline">
                {name}
              </Link>
              <div className="flex gap-2">
                <Button onClick={() => moveToCart.mutate(id)} disabled={moveToCart.isPending}>
                  Move to cart
                </Button>
                <Button
                  variant="outline"
                  onClick={() => {
                    removeWishlistProductId(id)
                    refresh()
                  }}
                >
                  Remove
                </Button>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

