"use client"

import {
  UserIcon,
  MapPinIcon,
  PackageIcon,
  BellIcon,
  MessageSquareIcon,
  StarIcon,
  ShieldIcon,
  SettingsIcon,
  MonitorIcon,
} from "lucide-react"
import { SidebarNav, type SidebarNavItem } from "@/components/shared/sidebar-nav"
import { routes } from "@/lib/routes"

const items: SidebarNavItem[] = [
  { href: routes.account.overview, label: "Overview", icon: UserIcon, exact: true },
  { href: routes.account.profile, label: "Profile", icon: UserIcon },
  { href: routes.account.addresses, label: "Addresses", icon: MapPinIcon },
  { href: routes.account.orders, label: "Orders", icon: PackageIcon },
  { href: routes.account.chat, label: "Chat", icon: MessageSquareIcon },
  { href: routes.account.reviews, label: "Reviews", icon: StarIcon },
  { href: routes.account.notifications, label: "Notifications", icon: BellIcon },
  { href: routes.account.sessions, label: "Sessions", icon: MonitorIcon },
  { href: routes.account.security, label: "Security", icon: ShieldIcon },
  { href: routes.account.preferences, label: "Preferences", icon: SettingsIcon },
]

export function AccountSidebar() {
  return <SidebarNav items={items} ariaLabel="Account" />
}
