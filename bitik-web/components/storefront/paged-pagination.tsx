"use client"

import * as React from "react"
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination"

function rangePages(current: number, total: number) {
  const out: Array<number | "ellipsis"> = []
  const push = (v: number | "ellipsis") => out.push(v)

  const window = 2
  const pages = new Set<number>([1, total])
  for (let p = current - window; p <= current + window; p++) {
    if (p >= 1 && p <= total) pages.add(p)
  }
  const sorted = Array.from(pages).sort((a, b) => a - b)
  for (let i = 0; i < sorted.length; i++) {
    const p = sorted[i]!
    const prev = sorted[i - 1]
    if (prev != null && p - prev > 1) push("ellipsis")
    push(p)
  }
  return out
}

export function PagedPagination({
  page,
  totalPages,
  hrefForPage,
}: {
  page: number
  totalPages: number
  hrefForPage: (page: number) => string
}) {
  const p = Math.max(1, page)
  const total = Math.max(1, totalPages)
  if (total <= 1) return null

  const pages = rangePages(p, total)

  return (
    <Pagination className="my-8">
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious
            href={hrefForPage(Math.max(1, p - 1))}
            aria-disabled={p <= 1}
            tabIndex={p <= 1 ? -1 : 0}
          />
        </PaginationItem>
        {pages.map((v, i) =>
          v === "ellipsis" ? (
            <PaginationItem key={`e-${i}`}>
              <PaginationEllipsis />
            </PaginationItem>
          ) : (
            <PaginationItem key={v}>
              <PaginationLink href={hrefForPage(v)} isActive={v === p}>
                {v}
              </PaginationLink>
            </PaginationItem>
          )
        )}
        <PaginationItem>
          <PaginationNext
            href={hrefForPage(Math.min(total, p + 1))}
            aria-disabled={p >= total}
            tabIndex={p >= total ? -1 : 0}
          />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  )
}

