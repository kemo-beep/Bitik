import Link from "next/link"
import { A11Y } from "@/lib/a11y"
import { routes } from "@/lib/routes"
import { GuestOnly } from "@/components/auth/guard"

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <GuestOnly>
      <div className="flex min-h-svh flex-col">
      <header
        role="banner"
        className="border-b"
      >
        <div className="mx-auto flex h-14 max-w-screen-2xl items-center px-4 lg:px-6">
          <Link
            href={routes.storefront.home}
            aria-label="Bitik home"
            className="font-heading text-base font-semibold tracking-tight"
          >
            Bitik
          </Link>
        </div>
      </header>
      <main
        id={A11Y.skipLinkId}
        role="main"
        className="flex flex-1 items-center justify-center px-4 py-10"
      >
        <div className="w-full max-w-sm">{children}</div>
      </main>
      </div>
    </GuestOnly>
  )
}
