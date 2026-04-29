"use client"

import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { sellerApi } from "@/lib/api/seller"
import { queryKeys } from "@/lib/queryKeys"
import { asArray } from "@/lib/safe"
import { SellerJsonView, SellerSectionCard } from "@/components/seller/section-card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

export function SellerProductEditorClient({ productId }: { productId?: string }) {
  const qc = useQueryClient()
  const [name, setName] = useState("")
  const [slug, setSlug] = useState("")
  const [description, setDescription] = useState("")
  const [imageURL, setImageURL] = useState("")

  const product = useQuery({
    queryKey: queryKeys.seller.product(productId ?? "new"),
    queryFn: () => sellerApi.getProduct(productId ?? ""),
    enabled: Boolean(productId),
  })
  const variants = useQuery({
    queryKey: queryKeys.seller.variants(productId ?? "new"),
    queryFn: () => sellerApi.listVariants(productId ?? ""),
    enabled: Boolean(productId),
  })
  const images = useQuery({
    queryKey: queryKeys.seller.productImages(productId ?? "new"),
    queryFn: () => sellerApi.listProductImages(productId ?? ""),
    enabled: Boolean(productId),
  })

  const refresh = () => {
    if (!productId) return
    qc.invalidateQueries({ queryKey: queryKeys.seller.product(productId) })
    qc.invalidateQueries({ queryKey: queryKeys.seller.variants(productId) })
    qc.invalidateQueries({ queryKey: queryKeys.seller.productImages(productId) })
  }

  const create = useMutation({ mutationFn: () => sellerApi.createProduct({ name, slug, description }) })
  const update = useMutation({
    mutationFn: () => sellerApi.patchProduct(productId ?? "", { name, slug, description }),
    onSuccess: refresh,
  })
  const addImage = useMutation({
    mutationFn: () => sellerApi.createProductImage(productId ?? "", { url: imageURL.trim() }),
    onSuccess: () => {
      setImageURL("")
      refresh()
    },
  })
  const addVariant = useMutation({
    mutationFn: () => sellerApi.createVariant(productId ?? "", { sku: `SKU-${Date.now()}`, name: "Default", price_cents: 1000 }),
    onSuccess: refresh,
  })

  return (
    <div className="space-y-4">
      <h1 className="font-heading text-2xl font-semibold">{productId ? "Edit product" : "New product"}</h1>
      <SellerSectionCard title="Product fields">
        <div className="grid gap-2 md:grid-cols-3">
          <Input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} />
          <Input placeholder="Slug" value={slug} onChange={(e) => setSlug(e.target.value)} />
          <Input placeholder="Description" value={description} onChange={(e) => setDescription(e.target.value)} />
        </div>
        {productId ? (
          <Button className="mt-3" onClick={() => update.mutate()} disabled={update.isPending}>
            Save changes
          </Button>
        ) : (
          <Button className="mt-3" onClick={() => create.mutate()} disabled={!name.trim() || create.isPending}>
            Create product
          </Button>
        )}
      </SellerSectionCard>

      {productId ? (
        <>
          <SellerSectionCard title="Images">
            <div className="flex gap-2">
              <Input placeholder="Image URL" value={imageURL} onChange={(e) => setImageURL(e.target.value)} />
              <Button onClick={() => addImage.mutate()} disabled={!imageURL.trim() || addImage.isPending}>Add</Button>
            </div>
            <SellerJsonView value={images.data} />
          </SellerSectionCard>
          <SellerSectionCard title="Variants / options">
            <Button onClick={() => addVariant.mutate()} disabled={addVariant.isPending}>Add default variant</Button>
            <SellerJsonView value={asArray((variants.data as Record<string, unknown> | undefined)?.items ?? variants.data) ?? variants.data} />
          </SellerSectionCard>
          <SellerSectionCard title="Current product payload">
            <SellerJsonView value={product.data} />
          </SellerSectionCard>
        </>
      ) : null}
    </div>
  )
}

