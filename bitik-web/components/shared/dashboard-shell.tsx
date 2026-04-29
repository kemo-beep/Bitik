import Link from "next/link"
import { cn } from "@/lib/utils"
import { A11Y } from "@/lib/a11y"
import { LanguageSwitcher } from "@/components/i18n/language-switcher"

export function DashboardShell({
  brand,
  brandHref,
  sidebar,
  children,
  topbar,
}: {
  brand: string
  brandHref: string
  sidebar: React.ReactNode
  children: React.ReactNode
  topbar?: React.ReactNode
}) {
  return (
    <div className="flex min-h-svh flex-col">
      <header
        role="banner"
        className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur"
      >
        <div className="mx-auto flex h-12 max-w-screen-2xl items-center gap-3 px-4 lg:px-6">
          <Link
            href={brandHref}
            className={cn("font-heading text-sm font-semibold tracking-tight")}
          >
            {brand}
          </Link>
          <div className="ml-auto flex items-center gap-2">
            <LanguageSwitcher />
            {topbar}
          </div>
        </div>
      </header>

      <div className="mx-auto flex w-full max-w-screen-2xl flex-1 gap-6 px-4 py-6 lg:px-6">
        <aside
          aria-label="Section navigation"
          className="sticky top-16 hidden h-fit w-56 shrink-0 md:block"
        >
          {sidebar}
        </aside>

        <main id={A11Y.skipLinkId} role="main" className="min-w-0 flex-1">
          {children}
        </main>
      </div>
    </div>
  )
}
