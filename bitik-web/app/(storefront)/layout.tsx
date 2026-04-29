import { Suspense } from "react"
import { StorefrontHeader } from "@/components/storefront/header"
import { StorefrontFooter } from "@/components/storefront/footer"
import { A11Y } from "@/lib/a11y"

export default function StorefrontLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="flex min-h-svh flex-col">
      <Suspense fallback={<div className="h-[148px] border-b bg-background md:h-[156px]" aria-hidden />}>
        <StorefrontHeader />
      </Suspense>
      <main id={A11Y.skipLinkId} role="main" className="flex-1">
        {children}
      </main>
      <StorefrontFooter />
    </div>
  )
}
