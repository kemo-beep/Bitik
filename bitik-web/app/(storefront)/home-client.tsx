"use client"

import * as React from "react"
import { useQuery } from "@tanstack/react-query"
import Image from "next/image"
import Link from "next/link"
import { ArrowRightIcon } from "lucide-react"
import { publicHome } from "@/lib/api/public"
import { queryKeys } from "@/lib/queryKeys"
import { routes } from "@/lib/routes"
import { asArray, asRecord, asString } from "@/lib/safe"
import { ProductCard } from "@/components/storefront/product-card"
import { Skeleton } from "@/components/ui/skeleton"
import { Carousel, CarouselContent, CarouselItem, type CarouselApi } from "@/components/ui/carousel"
import { cn } from "@/lib/utils"

const HERO_FALLBACKS = [
  {
    title: "Big Bachat Days",
    subtitle: "Sale is live",
    badge: "From ₹49",
    gradient: "from-sky-200 via-sky-100 to-blue-300",
    accent: "text-sky-900",
  },
  {
    title: "Quick fixes? Just DIY!",
    subtitle: "Level up your home",
    badge: "Up to 50% Off",
    gradient: "from-emerald-200 via-emerald-100 to-teal-200",
    accent: "text-emerald-900",
  },
  {
    title: "Furnish it your way!",
    subtitle: "Cushions, curtains, get it all",
    badge: "Up to 60% Off",
    gradient: "from-amber-200 via-orange-100 to-rose-200",
    accent: "text-amber-900",
  },
  {
    title: "Maximise your savings!",
    subtitle: "Avail GST benefits on home & furniture",
    badge: "Save more",
    gradient: "from-blue-300 via-blue-200 to-indigo-300",
    accent: "text-indigo-900",
  },
  {
    title: "Gifts from your heart…",
    subtitle: "to their home!",
    badge: "Up to 60% Off",
    gradient: "from-rose-200 via-pink-100 to-fuchsia-200",
    accent: "text-rose-900",
  },
  {
    title: "Repels mosquitoes",
    subtitle: "Uninterrupted air flow · Easy to install",
    badge: "Shop now",
    gradient: "from-violet-300 via-violet-200 to-purple-300",
    accent: "text-violet-900",
  },
  {
    title: "Tech that thrills",
    subtitle: "Latest mobiles, laptops, audio",
    badge: "Up to 70% Off",
    gradient: "from-yellow-200 via-amber-100 to-orange-200",
    accent: "text-amber-900",
  },
] as const

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
  const featured = products.slice(0, 4)

  const slides = (banners.length > 0 ? banners : HERO_FALLBACKS).map((b, i) => {
    const rec = asRecord(b)
    const fb = HERO_FALLBACKS[i % HERO_FALLBACKS.length]
    const rawLink = asString(rec?.link_url) ?? routes.storefront.products
    const link = rawLink.startsWith("/public/") ? rawLink.replace("/public", "") : rawLink
    return {
      title: asString(rec?.title) ?? fb.title,
      subtitle: asString(rec?.subtitle) ?? fb.subtitle,
      badge: asString(rec?.badge) ?? fb.badge,
      image: asString(rec?.image_url),
      link,
      gradient: fb.gradient,
      accent: fb.accent,
    }
  })

  const [carouselApi, setCarouselApi] = React.useState<CarouselApi>()
  const [active, setActive] = React.useState(0)
  const [paused, setPaused] = React.useState(false)
  const slideCount = slides.length

  React.useEffect(() => {
    if (!carouselApi) return
    const sync = () => setActive(carouselApi.selectedScrollSnap())
    sync()
    carouselApi.on("select", sync)
    carouselApi.on("reInit", sync)
    return () => {
      carouselApi.off("select", sync)
      carouselApi.off("reInit", sync)
    }
  }, [carouselApi])

  React.useEffect(() => {
    if (slideCount <= 1 || paused) return
    const id = window.setInterval(() => {
      setActive((p) => {
        const next = (p + 1) % slideCount
        carouselApi?.scrollTo(next)
        return next
      })
    }, 4000)
    return () => window.clearInterval(id)
  }, [carouselApi, slideCount, paused])

  React.useEffect(() => {
    carouselApi?.scrollTo(active)
  }, [active, carouselApi])

  return (
    <div className="w-full">
      <section className="mx-auto max-w-screen-2xl px-3 py-3 sm:px-4 lg:px-6">
        <Carousel
          setApi={setCarouselApi}
          opts={{ align: "start", loop: slideCount > 1, skipSnaps: false }}
          onMouseEnter={() => setPaused(true)}
          onMouseLeave={() => setPaused(false)}
          className="pb-1"
          role="region"
          aria-label="Promotions"
        >
          <CarouselContent className="-ml-3">
            {slides.map((s, i) => (
              <CarouselItem key={i} className="basis-full pl-3 sm:basis-1/2 lg:basis-1/3">
                <Link
                  href={s.link}
                  className={`group relative flex h-36 overflow-hidden rounded-lg bg-gradient-to-br ${s.gradient} sm:h-40`}
                >
                  <div className={`relative z-10 flex flex-1 flex-col justify-center gap-1 p-4 sm:p-5 ${s.accent}`}>
                    <h3 className="font-heading text-lg leading-tight font-bold tracking-tight sm:text-xl">
                      {s.title}
                    </h3>
                    <p className="text-xs font-medium opacity-80 sm:text-sm">{s.subtitle}</p>
                    {s.badge ? (
                      <div className="mt-1 inline-flex w-fit items-center rounded-full bg-white/70 px-2 py-0.5 text-xs font-semibold backdrop-blur">
                        {s.badge}
                      </div>
                    ) : null}
                    <div className="mt-1 inline-flex items-center gap-1 text-xs font-semibold underline-offset-4 group-hover:underline">
                      Shop now
                      <ArrowRightIcon className="size-3.5" />
                    </div>
                  </div>
                  {s.image ? (
                    <Image
                      src={s.image}
                      alt={s.title}
                      width={400}
                      height={300}
                      className="absolute inset-y-0 right-0 z-0 h-full w-2/5 object-cover opacity-95"
                      priority={i < 3}
                    />
                  ) : null}
                </Link>
              </CarouselItem>
            ))}
          </CarouselContent>
        </Carousel>

        {slideCount > 1 ? (
          <div className="mt-2 flex justify-center gap-1.5" role="tablist" aria-label="Slide navigation">
            {slides.map((_, i) => (
              <button
                key={i}
                type="button"
                role="tab"
                aria-selected={i === active}
                aria-label={`Go to slide ${i + 1}`}
                onClick={() => {
                  setActive(i)
                  carouselApi?.scrollTo(i)
                }}
                className={cn(
                  "h-1.5 rounded-full transition-all",
                  i === active ? "w-5 bg-foreground" : "w-1.5 bg-muted-foreground/40 hover:bg-muted-foreground/60"
                )}
              />
            ))}
          </div>
        ) : null}
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

        <div className="mt-4 grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
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
