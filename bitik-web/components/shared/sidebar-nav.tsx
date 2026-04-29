"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { cn } from "@/lib/utils"

export type SidebarNavItem = {
  href: string
  label: string
  icon?: React.ComponentType<{ className?: string }>
  exact?: boolean
}

export function SidebarNav({
  items,
  ariaLabel,
  className,
}: {
  items: SidebarNavItem[]
  ariaLabel: string
  className?: string
}) {
  const pathname = usePathname()
  return (
    <nav aria-label={ariaLabel} className={cn("flex flex-col gap-0.5", className)}>
      {items.map((item) => {
        const active = item.exact
          ? pathname === item.href
          : pathname === item.href || pathname.startsWith(item.href + "/")
        const Icon = item.icon
        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={active ? "page" : undefined}
            className={cn(
              "inline-flex items-center gap-2 rounded-md px-2.5 py-1.5 text-sm transition-colors",
              "text-muted-foreground hover:bg-muted hover:text-foreground",
              "aria-[current=page]:bg-muted aria-[current=page]:text-foreground"
            )}
          >
            {Icon ? <Icon className="size-4" /> : null}
            <span>{item.label}</span>
          </Link>
        )
      })}
    </nav>
  )
}
