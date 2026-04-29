"use client"

import {
  LayoutDashboardIcon,
  StoreIcon,
  PackageIcon,
  BoxesIcon,
  ShoppingBagIcon,
  TruckIcon,
  TagIcon,
  StarIcon,
  WalletIcon,
  BarChart3Icon,
  SettingsIcon,
  BellIcon,
  MessageSquareIcon,
} from "lucide-react"
import { SidebarNav, type SidebarNavItem } from "@/components/shared/sidebar-nav"
import { routes } from "@/lib/routes"

const items: SidebarNavItem[] = [
  { href: routes.seller.dashboard, label: "Dashboard", icon: LayoutDashboardIcon, exact: true },
  { href: routes.seller.products, label: "Products", icon: PackageIcon },
  { href: routes.seller.inventory, label: "Inventory", icon: BoxesIcon },
  { href: routes.seller.orders, label: "Orders", icon: ShoppingBagIcon },
  { href: routes.seller.chat, label: "Chat", icon: MessageSquareIcon },
  { href: routes.seller.shipping, label: "Shipping", icon: TruckIcon },
  { href: "/seller/notifications", label: "Notifications", icon: BellIcon },
  { href: routes.seller.promotions, label: "Promotions", icon: TagIcon },
  { href: routes.seller.reviews, label: "Reviews", icon: StarIcon },
  { href: routes.seller.wallet, label: "Wallet", icon: WalletIcon },
  { href: routes.seller.analytics, label: "Analytics", icon: BarChart3Icon },
  { href: routes.seller.profile, label: "Shop", icon: StoreIcon },
  { href: routes.seller.settings, label: "Settings", icon: SettingsIcon },
]

export function SellerSidebar() {
  return <SidebarNav items={items} ariaLabel="Seller center" />
}
