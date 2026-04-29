"use client"

import * as React from "react"
import Image from "next/image"
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

export type GalleryImage = {
  src: string
  alt: string
  caption?: string
}

export function ImageGallery({
  images,
  initial = 0,
  className,
  thumbnails = true,
}: {
  images: GalleryImage[]
  initial?: number
  className?: string
  thumbnails?: boolean
}) {
  const [active, setActive] = React.useState(Math.min(initial, Math.max(0, images.length - 1)))
  const total = images.length
  const main = images[active]

  const go = React.useCallback(
    (delta: number) => {
      if (total === 0) return
      setActive((i) => (i + delta + total) % total)
    },
    [total]
  )

  function onKeyDown(e: React.KeyboardEvent<HTMLDivElement>) {
    if (e.key === "ArrowLeft") {
      e.preventDefault()
      go(-1)
    } else if (e.key === "ArrowRight") {
      e.preventDefault()
      go(1)
    }
  }

  if (total === 0) {
    return (
      <div
        className={cn(
          "flex aspect-square w-full items-center justify-center rounded-xl border border-dashed bg-muted/40 text-sm text-muted-foreground",
          className
        )}
        role="img"
        aria-label="No images"
      >
        No images
      </div>
    )
  }

  return (
    <div
      className={cn("flex flex-col gap-3", className)}
      role="region"
      aria-roledescription="image gallery"
      aria-label="Product images"
      tabIndex={0}
      onKeyDown={onKeyDown}
    >
      <div className="relative aspect-square w-full overflow-hidden rounded-xl border bg-muted">
        <Image
          src={main.src}
          alt={main.alt}
          fill
          sizes="(max-width: 768px) 100vw, 50vw"
          className="object-cover"
          draggable={false}
        />
        {total > 1 ? (
          <>
            <Button
              variant="secondary"
              size="icon-sm"
              onClick={() => go(-1)}
              aria-label="Previous image"
              className="absolute top-1/2 left-2 -translate-y-1/2 rounded-full"
            >
              <ChevronLeftIcon />
            </Button>
            <Button
              variant="secondary"
              size="icon-sm"
              onClick={() => go(1)}
              aria-label="Next image"
              className="absolute top-1/2 right-2 -translate-y-1/2 rounded-full"
            >
              <ChevronRightIcon />
            </Button>
            <span
              className="absolute bottom-2 right-2 rounded-full bg-background/80 px-2 py-0.5 text-xs"
              aria-live="polite"
            >
              {active + 1} / {total}
            </span>
          </>
        ) : null}
      </div>

      {thumbnails && total > 1 ? (
        <ul
          className="flex gap-2 overflow-x-auto"
          role="tablist"
          aria-label="Image thumbnails"
        >
          {images.map((img, i) => (
            <li key={i}>
              <button
                type="button"
                role="tab"
                aria-selected={i === active}
                aria-label={`Show image ${i + 1}`}
                onClick={() => setActive(i)}
                className={cn(
                  "relative size-16 shrink-0 overflow-hidden rounded-md border bg-muted outline-none transition-colors",
                  "focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/40",
                  i === active && "border-primary ring-2 ring-primary/30"
                )}
              >
                <Image src={img.src} alt="" fill sizes="64px" className="object-cover" />
              </button>
            </li>
          ))}
        </ul>
      ) : null}

      {main.caption ? (
        <p className="text-sm text-muted-foreground">{main.caption}</p>
      ) : null}
    </div>
  )
}
