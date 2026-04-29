import Link from "next/link"
import { routes } from "@/lib/routes"

const sections = [
  {
    title: "Shop",
    links: [
      { label: "Categories", href: routes.storefront.categories },
      { label: "Brands", href: routes.storefront.brands },
      { label: "All products", href: routes.storefront.products },
      { label: "Search", href: routes.storefront.search },
    ],
  },
  {
    title: "Account",
    links: [
      { label: "Sign in", href: routes.auth.login },
      { label: "Register", href: routes.auth.register },
      { label: "Orders", href: routes.account.orders },
      { label: "Wishlist", href: routes.storefront.wishlist },
    ],
  },
  {
    title: "Sell on Bitik",
    links: [
      { label: "Apply as seller", href: routes.seller.apply },
      { label: "Seller center", href: routes.seller.dashboard },
    ],
  },
  {
    title: "Help",
    links: [
      { label: "FAQ", href: routes.storefront.pages("faq") },
      { label: "Shipping", href: routes.storefront.pages("shipping") },
      { label: "Returns", href: routes.storefront.pages("returns") },
      { label: "Contact", href: routes.storefront.pages("contact") },
    ],
  },
]

export function StorefrontFooter() {
  return (
    <footer
      role="contentinfo"
      className="mt-16 border-t bg-muted/30"
    >
      <div className="mx-auto grid max-w-screen-2xl grid-cols-2 gap-8 px-4 py-10 md:grid-cols-4 lg:px-6">
        {sections.map((s) => (
          <nav key={s.title} aria-label={s.title} className="flex flex-col gap-2">
            <h2 className="text-sm font-semibold">{s.title}</h2>
            <ul className="flex flex-col gap-1.5 text-sm text-muted-foreground">
              {s.links.map((l) => (
                <li key={l.href}>
                  <Link href={l.href} className="hover:text-foreground">
                    {l.label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>
        ))}
      </div>
      <div className="border-t">
        <div className="mx-auto flex max-w-screen-2xl flex-col items-start justify-between gap-2 px-4 py-4 text-xs text-muted-foreground sm:flex-row sm:items-center lg:px-6">
          <p>© {new Date().getFullYear()} Bitik. All rights reserved.</p>
          <div className="flex gap-3">
            <Link href={routes.storefront.pages("terms")} className="hover:text-foreground">Terms</Link>
            <Link href={routes.storefront.pages("privacy")} className="hover:text-foreground">Privacy</Link>
          </div>
        </div>
      </div>
    </footer>
  )
}
