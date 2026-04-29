"use client"

import { useQuery } from "@tanstack/react-query"
import Image from "next/image"
import Link from "next/link"
import { listCategories } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { routes } from "@/lib/routes"
import { asArray, asRecord, asString } from "@/lib/safe"
import { Skeleton } from "@/components/ui/skeleton"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export function CategoriesClient() {
  const q = useQuery({
    queryKey: queryKeys.public.categories(),
    queryFn: listCategories,
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    retry: 1,
  })

  return (
    <div className="mx-auto w-full max-w-screen-2xl px-4 py-8 lg:px-6">
      <header className="flex flex-col gap-2">
        <h1 className="font-heading text-2xl font-semibold tracking-tight">Categories</h1>
        <p className="text-sm text-muted-foreground">Browse categories in the public catalog.</p>
      </header>

      <section className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {q.isLoading ? (
          Array.from({ length: 12 }).map((_, i) => <Skeleton key={i} className="h-36 rounded-xl" />)
        ) : q.isError ? (
          <div className="col-span-full rounded-xl border bg-destructive/5 p-6 text-sm">
            Could not load categories.
          </div>
        ) : (
          (asArray(q.data) ?? []).map((c, i) => {
            const rec = asRecord(c)
            const id = asString(rec?.id)
            const name = asString(rec?.name) ?? "Unnamed category"
            const imageUrl = asString(rec?.image_url)
            const href = id ? routes.storefront.category(id) : routes.storefront.categories
            return (
              <Card key={id ?? String(i)} className="overflow-hidden">
                {imageUrl ? (
                  <Image
                    src={imageUrl}
                    alt=""
                    width={800}
                    height={500}
                    className="h-36 w-full object-cover"
                  />
                ) : null}
                <CardHeader>
                  <CardTitle className="line-clamp-1">
                    <Link href={href} className="hover:underline">
                      {name}
                    </Link>
                  </CardTitle>
                </CardHeader>
                <CardContent className="text-xs text-muted-foreground">Explore products →</CardContent>
              </Card>
            )
          })
        )}
      </section>
    </div>
  )
}

