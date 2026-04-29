"use client"

import Link from "next/link"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { asArray, asString } from "@/lib/safe"
import { SellerSectionCard } from "@/components/seller/section-card"
import { Button } from "@/components/ui/button"
import { routes } from "@/lib/routes"
import { useAnalytics, analyticsEvents } from "@/lib/analytics"

export function SellerProductsClient() {
  const qc = useQueryClient()
  const products = useQuery({ queryKey: queryKeys.seller.products(), queryFn: () => sellerApi.listProducts() })
  const refresh = () => qc.invalidateQueries({ queryKey: queryKeys.seller.products() })
  const publish = useMutation({ mutationFn: (id: string) => sellerApi.publishProduct(id), onSuccess: refresh })
  const unpublish = useMutation({ mutationFn: (id: string) => sellerApi.unpublishProduct(id), onSuccess: refresh })
  const duplicate = useMutation({ mutationFn: (id: string) => sellerApi.duplicateProduct(id), onSuccess: refresh })
  const remove = useMutation({ mutationFn: (id: string) => sellerApi.deleteProduct(id), onSuccess: refresh })
  const items = asArray((products.data as Record<string, unknown> | undefined)?.items ?? products.data) ?? []
  const analytics = useAnalytics()

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="font-heading text-2xl font-semibold">Products</h1>
        <Button
          nativeButton={false}
          render={
            <Link
              href={routes.seller.productNew}
              onClick={() => analytics.track({ name: analyticsEvents.sellerProductCreate })}
            >
              New product
            </Link>
          }
        />
      </div>
      <SellerSectionCard title="Product list" description="Filter/status controls can be layered on top of this list.">
        <div className="space-y-3">
          {products.isLoading ? <div className="text-sm">Loading products…</div> : null}
          {items.map((item, i) => {
            const rec = item as Record<string, unknown>
            const id = asString(rec.id) ?? `${i}`
            const name = asString(rec.name) ?? "Product"
            return (
              <div key={id} className="rounded border p-3">
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <Link className="font-medium hover:underline" href={routes.seller.product(id)}>
                    {name}
                  </Link>
                  <div className="flex flex-wrap gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => {
                        analytics.track({ name: analyticsEvents.sellerProductPublish, properties: { product_id: id } })
                        publish.mutate(id)
                      }}
                    >
                      Publish
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => unpublish.mutate(id)}>Unpublish</Button>
                    <Button size="sm" variant="outline" onClick={() => duplicate.mutate(id)}>Duplicate</Button>
                    <Button size="sm" variant="destructive" onClick={() => remove.mutate(id)}>Delete</Button>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      </SellerSectionCard>
    </div>
  )
}

