"use client"

import * as React from "react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  SearchIcon,
  ShoppingCartIcon,
  HeartIcon,
  CircleUserRoundIcon,
  BellIcon,
  ChevronDownIcon,
  MapPinIcon,
  StoreIcon,
  SparklesIcon,
  ShirtIcon,
  SmartphoneIcon,
  LaptopIcon,
  BrushIcon,
  HomeIcon,
  TvIcon,
  BabyIcon,
  UtensilsCrossedIcon,
  HardHatIcon,
  BikeIcon,
  TrophyIcon,
  BookIcon,
  ArmchairIcon,
} from "lucide-react"
import { Input } from "@/components/ui/input"
import { routes } from "@/lib/routes"
import { UnreadIndicator } from "@/components/notifications/unread-indicator"
import { useI18n } from "@/lib/i18n/context"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

export function StorefrontHeader() {
  const { t, locale, setLocale } = useI18n()
  const pathname = usePathname()
  const [qParam, setQParam] = React.useState<string | null>(null)
  React.useEffect(() => {
    setQParam(new URLSearchParams(window.location.search).get("q"))
  }, [pathname])

  const stripItems = [
    { href: "/", label: t("nav.forYou"), icon: SparklesIcon },
    { href: routes.storefront.categories, label: t("nav.fashion"), icon: ShirtIcon },
    { href: `${routes.storefront.search}?q=mobile`, label: t("nav.mobiles"), icon: SmartphoneIcon },
    { href: `${routes.storefront.search}?q=beauty`, label: t("nav.beauty"), icon: BrushIcon },
    { href: `${routes.storefront.search}?q=electronics`, label: t("nav.electronics"), icon: LaptopIcon },
    { href: `${routes.storefront.search}?q=home`, label: t("nav.home"), icon: HomeIcon },
    { href: `${routes.storefront.search}?q=appliances`, label: t("nav.appliances"), icon: TvIcon },
    { href: `${routes.storefront.search}?q=toys`, label: t("nav.toys"), icon: BabyIcon },
    { href: `${routes.storefront.search}?q=food`, label: t("nav.food"), icon: UtensilsCrossedIcon },
    { href: `${routes.storefront.search}?q=auto`, label: t("nav.auto"), icon: HardHatIcon },
    { href: `${routes.storefront.search}?q=2wheelers`, label: t("nav.twoWheelers"), icon: BikeIcon },
    { href: `${routes.storefront.search}?q=sports`, label: t("nav.sports"), icon: TrophyIcon },
    { href: `${routes.storefront.search}?q=books`, label: t("nav.books"), icon: BookIcon },
    { href: `${routes.storefront.search}?q=furniture`, label: t("nav.furniture"), icon: ArmchairIcon },
  ] as const

  return (
    <header
      role="banner"
      className="sticky top-0 z-50 border-b border-border/80 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/90"
    >
      <div className="mx-auto max-w-screen-2xl">
        {/* Top: logo + location */}
        <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-b border-border/50 px-3 py-2 sm:px-4 lg:px-6">
          <Link
            href={routes.storefront.home}
            aria-label="Bitik home"
            className="inline-flex min-w-0 shrink-0 items-center gap-1.5 font-heading text-base font-semibold tracking-tight"
          >
            <span className="grid size-7 shrink-0 place-items-center rounded-sm bg-foreground text-background">
              <StoreIcon className="size-4" />
            </span>
            Bitik
          </Link>
          <div className="hidden items-center gap-1.5 text-sm text-foreground sm:flex">
            <MapPinIcon className="size-4 shrink-0 text-foreground" aria-hidden />
            <span className="text-muted-foreground">{t("nav.locationUnset")}</span>
            <Link
              href={routes.account.addresses}
              className="font-medium text-[#2874F0] hover:underline dark:text-sky-400"
            >
              {t("nav.selectLocation")}
            </Link>
          </div>
        </div>

        {/* Search + account */}
        <div className="flex flex-wrap items-center gap-3 px-3 py-2.5 sm:gap-4 sm:px-4 lg:px-6">
          <form
            role="search"
            action={routes.storefront.search}
            className="order-2 min-w-0 flex-1 md:order-1 md:max-w-2xl"
          >
            <label htmlFor="storefront-search" className="sr-only">
              {t("nav.searchLabel")}
            </label>
            <div className="relative">
              <SearchIcon className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="storefront-search"
                name="q"
                type="search"
                placeholder={t("nav.searchPlaceholder")}
                className="h-11 w-full rounded-full border-2 border-sky-300/90 bg-background pl-10 pr-4 text-base shadow-sm placeholder:text-muted-foreground focus-visible:border-sky-500 focus-visible:ring-sky-500/30 md:text-sm dark:border-sky-600/50 dark:focus-visible:border-sky-500"
              />
            </div>
          </form>

          <div className="order-1 flex w-full items-center justify-end gap-0.5 sm:order-2 sm:w-auto sm:shrink-0 sm:gap-1 md:ms-auto">
            <Link
              aria-label={t("nav.searchLabel")}
              href={routes.storefront.search}
              className="inline-flex size-10 items-center justify-center rounded-full hover:bg-muted md:hidden"
            >
              <SearchIcon className="size-5" />
            </Link>

            <DropdownMenu>
              <DropdownMenuTrigger className="inline-flex items-center gap-0.5 rounded-md px-2 py-2 text-sm font-medium text-foreground hover:bg-muted data-popup-open:bg-muted">
                <CircleUserRoundIcon className="size-5" aria-hidden />
                <span className="hidden sm:inline">{t("nav.login")}</span>
                <ChevronDownIcon className="size-3.5 opacity-70" aria-hidden />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="min-w-44">
                <DropdownMenuItem
                  nativeButton={false}
                  render={
                    <Link href={routes.auth.login} className="flex cursor-default items-center gap-1.5">
                      {t("nav.login")}
                    </Link>
                  }
                />
                <DropdownMenuItem
                  nativeButton={false}
                  render={
                    <Link href={routes.auth.register} className="flex cursor-default items-center gap-1.5">
                      {t("nav.register")}
                    </Link>
                  }
                />
              </DropdownMenuContent>
            </DropdownMenu>

            <DropdownMenu>
              <DropdownMenuTrigger className="inline-flex items-center gap-0.5 rounded-md px-2 py-2 text-sm font-medium text-foreground hover:bg-muted data-popup-open:bg-muted">
                <span>{t("nav.more")}</span>
                <ChevronDownIcon className="size-3.5 opacity-70" aria-hidden />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="min-w-48">
                <DropdownMenuItem
                  nativeButton={false}
                  render={
                    <Link
                      href={routes.storefront.wishlist}
                      className="flex cursor-default items-center gap-1.5"
                    >
                      <HeartIcon className="size-4" />
                      {t("nav.wishlist")}
                    </Link>
                  }
                />
                <DropdownMenuItem
                  nativeButton={false}
                  render={
                    <Link
                      href={routes.account.notifications}
                      className="flex cursor-default items-center gap-1.5"
                    >
                      <BellIcon className="size-4" />
                      {t("nav.notifications")}
                    </Link>
                  }
                />
                <DropdownMenuSeparator />
                <DropdownMenuItem closeOnClick={false} onClick={() => setLocale("en")}>
                  English {locale === "en" ? "✓" : ""}
                </DropdownMenuItem>
                <DropdownMenuItem closeOnClick={false} onClick={() => setLocale("fr")}>
                  Français {locale === "fr" ? "✓" : ""}
                </DropdownMenuItem>
                <DropdownMenuItem closeOnClick={false} onClick={() => setLocale("ar")}>
                  العربية {locale === "ar" ? "✓" : ""}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>

            <div className="flex items-center gap-1">
              <UnreadIndicator scope="buyer" />
              <Link
                href={routes.storefront.cart}
                className="inline-flex items-center gap-1.5 rounded-md px-2 py-2 text-sm font-medium text-foreground hover:bg-muted"
              >
                <ShoppingCartIcon className="size-5" aria-hidden />
                <span className="hidden lg:inline">{t("nav.cart")}</span>
              </Link>
            </div>
          </div>
        </div>

        {/* Category-style strip */}
        <nav
          aria-label={t("nav.categoryNav")}
          className="no-scrollbar flex overflow-x-auto border-t border-border/50 px-2 py-1 sm:px-4 lg:px-6"
        >
          <ul className="flex min-w-max items-stretch gap-0.5 px-1 py-1 sm:gap-1.5">
            {stripItems.map(({ href, label, icon: Icon }) => {
              const wantQ = href.includes("?") ? new URL(href, "https://example.com").searchParams.get("q") : null
              const active =
                href === "/"
                  ? pathname === "/"
                  : wantQ != null
                    ? pathname.startsWith("/search") && qParam === wantQ
                    : pathname === href || pathname.startsWith(`${href}/`)

              return (
                <li key={href}>
                  <Link
                    href={href}
                    className={cn(
                      "flex min-w-[4rem] flex-col items-center gap-1 rounded-md px-1 py-1 text-center text-[11px] font-medium text-muted-foreground transition-colors hover:text-foreground sm:min-w-[4.75rem] sm:px-1.5 sm:text-sm",
                      active &&
                        "text-foreground after:mx-auto after:mt-0.5 after:block after:h-1 after:w-10 after:rounded-full after:bg-[#2874F0] after:content-[''] dark:after:bg-sky-400"
                    )}
                  >
                    <span
                      className={cn(
                        "grid size-10 place-items-center rounded-md border border-border/60 bg-muted/20 text-muted-foreground transition-colors",
                        active
                          ? "border-cyan-500/35 bg-cyan-500/8 text-foreground dark:border-cyan-400/35 dark:bg-cyan-400/10"
                          : "hover:border-border hover:bg-muted/40"
                      )}
                    >
                      <Icon
                        className={cn(
                          "size-5 shrink-0 transition-colors",
                          active
                            ? "text-amber-400 dark:text-amber-300"
                            : "text-amber-500/90 dark:text-amber-300/90"
                        )}
                        strokeWidth={1.75}
                        aria-hidden
                      />
                    </span>
                    <span className={cn("max-w-[5rem] truncate", active && "font-semibold text-foreground")}>
                      {label}
                    </span>
                  </Link>
                </li>
              )
            })}
          </ul>
        </nav>
      </div>
    </header>
  )
}
