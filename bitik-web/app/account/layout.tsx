import { StorefrontHeader } from "@/components/storefront/header"
import { StorefrontFooter } from "@/components/storefront/footer"
import { AccountSidebar } from "@/components/account/sidebar"
import { A11Y } from "@/lib/a11y"
import { RequireRole } from "@/components/auth/guard"

export default function AccountLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <RequireRole role="buyer">
      <div className="flex min-h-svh flex-col">
      <StorefrontHeader />
      <div className="mx-auto flex w-full max-w-screen-2xl flex-1 gap-6 px-4 py-6 lg:px-6">
        <aside
          aria-label="Account navigation"
          className="hidden w-56 shrink-0 md:block"
        >
          <AccountSidebar />
        </aside>
        <main id={A11Y.skipLinkId} role="main" className="min-w-0 flex-1">
          {children}
        </main>
      </div>
      <StorefrontFooter />
      </div>
    </RequireRole>
  )
}
