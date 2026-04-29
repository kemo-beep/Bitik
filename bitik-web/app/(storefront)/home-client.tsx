"use client"

import { useQuery } from "@tanstack/react-query"
import Image from "next/image"
import Link from "next/link"
import {
  ArrowRightIcon,
  SearchIcon,
  ShieldCheckIcon,
  TruckIcon,
  WalletCardsIcon,
} from "lucide-react"
import { publicHome } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { routes } from "@/lib/routes"
import { asArray, asRecord, asString } from "@/lib/safe"
import { ProductCard } from "@/components/storefront/product-card"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"

export function HomeClient() {
  const q = useQuery({
    queryKey: queryKeys.public.home(),
    queryFn: publicHome,
    staleTime: 60_000,
    gcTime: 10 * 60_000,
    retry: 1,
  })

  const data = asRecord(q.data) ?? {}
  const banners = asArray(data.banners) ?? []
  const products = asArray(data.products) ?? []
  const banner0 = asRecord(banners[0])
  const heroImage = asString(banner0?.image_url)
  const heroTitle = asString(banner0?.title) ?? "Shop Bitik"
  const rawHeroLink = asString(banner0?.link_url) ?? routes.storefront.products
  const heroLink = rawHeroLink.startsWith("/public/")
    ? rawHeroLink.replace("/public", "")
    : rawHeroLink
  const featured = products.slice(0, 4)

  return (
    <div className="w-full">
      <section className="relative border-b bg-foreground text-background">
        <div className="absolute inset-0">
          {heroImage ? (
            <Image
              src={heroImage}
              alt={heroTitle}
              width={1600}
              height={800}
              priority
              className="h-full w-full object-cover opacity-55"
            />
          ) : (
            <div className="h-full w-full bg-[radial-gradient(circle_at_70%_30%,rgba(255,255,255,.22),transparent_32%),linear-gradient(135deg,#171717,#3f1d13_55%,#111)]" />
          )}
          <div className="absolute inset-0 bg-gradient-to-r from-black/80 via-black/45 to-black/10" />
        </div>

        <div className="relative mx-auto grid min-h-[420px] max-w-screen-2xl gap-8 px-4 py-10 sm:min-h-[520px] sm:px-6 lg:grid-cols-[minmax(0,0.95fr)_minmax(360px,0.65fr)] lg:items-end lg:px-8">
          <div className="max-w-2xl animate-in duration-700 fade-in slide-in-from-bottom-3">
            <div className="mb-3 text-xs font-medium tracking-[0.22em] text-background/70 uppercase">
              Bitik marketplace
            </div>
            <h1 className="font-heading text-4xl leading-[0.95] font-semibold tracking-tight sm:text-6xl lg:text-7xl">
              {heroTitle}
            </h1>
            <p className="mt-4 max-w-lg text-sm leading-6 text-background/75 sm:text-base">
              Curated everyday goods, trusted sellers, and checkout built for
              local commerce.
            </p>
            <div className="mt-6 flex flex-wrap gap-2">
              <Button
                className="bg-background text-foreground hover:bg-background/90"
                nativeButton={false}
                render={
                  <Link href={heroLink}>
                    Shop now
                    <ArrowRightIcon data-icon="inline-end" />
                  </Link>
                }
              />
              <Button
                variant="outline"
                className="border-background/35 bg-transparent text-background hover:bg-background/10 hover:text-background"
                nativeButton={false}
                render={
                  <Link href={routes.storefront.search}>
                    <SearchIcon data-icon="inline-start" />
                    Search
                  </Link>
                }
              />
            </div>
          </div>

          <div className="hidden border-y border-background/20 py-4 text-sm text-background/80 lg:grid lg:grid-cols-3">
            <div className="flex items-center gap-2 pr-4">
              <ShieldCheckIcon className="size-4" />
              Buyer protection
            </div>
            <div className="flex items-center gap-2 border-x border-background/20 px-4">
              <TruckIcon className="size-4" />
              Local delivery
            </div>
            <div className="flex items-center gap-2 pl-4">
              <WalletCardsIcon className="size-4" />
              Wave and POD
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto grid max-w-screen-2xl gap-3 px-4 py-5 sm:grid-cols-3 lg:px-6">
        <Link
          href={routes.storefront.products}
          className="rounded-md border p-3 text-sm transition-colors hover:border-foreground/25"
        >
          <div className="font-medium">All products</div>
          <div className="mt-1 text-xs text-muted-foreground">
            Browse the catalog
          </div>
        </Link>
        <Link
          href={routes.storefront.categories}
          className="rounded-md border p-3 text-sm transition-colors hover:border-foreground/25"
        >
          <div className="font-medium">Categories</div>
          <div className="mt-1 text-xs text-muted-foreground">Shop by need</div>
        </Link>
        <Link
          href={routes.storefront.brands}
          className="rounded-md border p-3 text-sm transition-colors hover:border-foreground/25"
        >
          <div className="font-medium">Brands</div>
          <div className="mt-1 text-xs text-muted-foreground">
            Find trusted names
          </div>
        </Link>
      </section>

      <section className="mx-auto max-w-screen-2xl px-4 py-6 lg:px-6">
        <div className="flex items-end justify-between gap-4">
          <div>
            <h2 className="font-heading text-xl font-semibold tracking-tight">
              Popular now
            </h2>
            <p className="text-sm text-muted-foreground">
              Fast-moving picks from active sellers.
            </p>
          </div>
          <Link
            href={routes.storefront.products}
            className="inline-flex items-center gap-1 text-sm font-medium hover:underline"
          >
            View all <ArrowRightIcon className="size-3.5" />
          </Link>
        </div>

        {featured.length > 0 && !q.isLoading && !q.isError ? (
          <div className="mt-4 grid gap-3 border-y py-4 md:grid-cols-4">
            {featured.map((p, i) => {
              const rec = asRecord(p)
              return (
                <Link
                  key={String(i)}
                  href={
                    asString(rec?.slug)
                      ? routes.storefront.productSlug(asString(rec?.slug) ?? "")
                      : routes.storefront.products
                  }
                  className="text-sm font-medium hover:underline"
                >
                  {asString(rec?.name)}
                </Link>
              )
            })}
          </div>
        ) : null}

        <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6">
          {q.isLoading ? (
            Array.from({ length: 12 }).map((_, i) => (
              <Skeleton key={i} className="aspect-[4/5] rounded-md" />
            ))
          ) : q.isError ? (
            <div className="col-span-full rounded-md border bg-destructive/5 p-4 text-sm">
              Could not load the homepage.
            </div>
          ) : (
            products
              .slice(0, 12)
              .map((p, i) => <ProductCard key={String(i)} product={p} />)
          )}
        </div>
      </section>
    </div>
  )
}
