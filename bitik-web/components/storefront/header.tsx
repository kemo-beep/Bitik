"use client"

import Link from "next/link"
import {
  SearchIcon,
  ShoppingCartIcon,
  HeartIcon,
  UserRoundIcon,
  BellIcon,
  StoreIcon,
} from "lucide-react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { routes } from "@/lib/routes"
import { UnreadIndicator } from "@/components/notifications/unread-indicator"
import { LanguageSwitcher } from "@/components/i18n/language-switcher"
import { useI18n } from "@/lib/i18n/context"

export function StorefrontHeader() {
  const { t } = useI18n()
  return (
    <header
      role="banner"
      className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80"
    >
      <div className="mx-auto flex h-14 max-w-screen-2xl items-center gap-2 px-3 sm:gap-3 sm:px-4 lg:px-6">
        <Link
          href={routes.storefront.home}
          aria-label="Bitik home"
          className="inline-flex items-center gap-1.5 font-heading text-base font-semibold tracking-tight"
        >
          <span className="grid size-7 place-items-center rounded-sm bg-foreground text-background">
            <StoreIcon className="size-4" />
          </span>
          Bitik
        </Link>

        <nav
          aria-label="Primary"
          className="hidden items-center gap-1 text-sm md:flex"
        >
          <Link
            className="px-2 py-1 text-muted-foreground hover:text-foreground"
            href={routes.storefront.categories}
          >
            {t("nav.categories")}
          </Link>
          <Link
            className="px-2 py-1 text-muted-foreground hover:text-foreground"
            href={routes.storefront.brands}
          >
            {t("nav.brands")}
          </Link>
          <Link
            className="px-2 py-1 text-muted-foreground hover:text-foreground"
            href={routes.storefront.products}
          >
            {t("nav.products")}
          </Link>
        </nav>

        <form
          role="search"
          action={routes.storefront.search}
          className="ml-auto hidden max-w-xl flex-1 md:block"
        >
          <label htmlFor="storefront-search" className="sr-only">
            {t("nav.searchLabel")}
          </label>
          <div className="relative">
            <SearchIcon className="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="storefront-search"
              name="q"
              type="search"
              placeholder={t("nav.searchPlaceholder")}
              className="pl-8"
            />
          </div>
        </form>

        <div className="ml-auto flex items-center gap-1 md:ml-0">
          <Link
            aria-label="Search"
            href={routes.storefront.search}
            className="inline-flex size-8 items-center justify-center rounded-md hover:bg-muted md:hidden"
          >
            <SearchIcon className="size-4" />
          </Link>
          <LanguageSwitcher />
          <Link
            aria-label="Wishlist"
            href={routes.storefront.wishlist}
            className="inline-flex size-8 items-center justify-center rounded-md hover:bg-muted"
          >
            <HeartIcon className="size-4" />
          </Link>
          <Link
            aria-label="Notifications"
            href={routes.account.notifications}
            className="inline-flex size-8 items-center justify-center rounded-md hover:bg-muted"
          >
            <BellIcon className="size-4" />
          </Link>
          <UnreadIndicator scope="buyer" />
          <Link
            aria-label="Cart"
            href={routes.storefront.cart}
            className="inline-flex size-8 items-center justify-center rounded-md hover:bg-muted"
          >
            <ShoppingCartIcon className="size-4" />
          </Link>
          <Button
            variant="outline"
            size="sm"
            nativeButton={false}
            render={
              <Link href={routes.auth.login} aria-label="Sign in">
                <UserRoundIcon data-icon="inline-start" />
                {t("nav.signIn")}
              </Link>
            }
          />
        </div>
      </div>
    </header>
  )
}
