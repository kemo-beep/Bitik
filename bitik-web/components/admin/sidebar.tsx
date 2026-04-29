"use client"

import {
  LayoutDashboardIcon,
  UsersIcon,
  StoreIcon,
  PackageIcon,
  ShoppingBagIcon,
  CreditCardIcon,
  TruckIcon,
  TagIcon,
  ShieldAlertIcon,
  FileTextIcon,
  KeyRoundIcon,
  SettingsIcon,
  ScrollTextIcon,
  ActivityIcon,
  BellIcon,
} from "lucide-react"
import { SidebarNav, type SidebarNavItem } from "@/components/shared/sidebar-nav"
import { routes } from "@/lib/routes"

const items: SidebarNavItem[] = [
  { href: routes.admin.dashboard, label: "Dashboard", icon: LayoutDashboardIcon, exact: true },
  { href: routes.admin.users, label: "Users", icon: UsersIcon },
  { href: routes.admin.sellers, label: "Sellers", icon: StoreIcon },
  { href: routes.admin.products, label: "Products", icon: PackageIcon },
  { href: routes.admin.categories, label: "Categories", icon: TagIcon },
  { href: routes.admin.brands, label: "Brands", icon: TagIcon },
  { href: routes.admin.orders, label: "Orders", icon: ShoppingBagIcon },
  { href: routes.admin.payments, label: "Payments", icon: CreditCardIcon },
  { href: routes.admin.shipments, label: "Shipments", icon: TruckIcon },
  { href: "/admin/notifications", label: "Notifications", icon: BellIcon },
  { href: routes.admin.promotions, label: "Promotions", icon: TagIcon },
  { href: routes.admin.moderation, label: "Moderation", icon: ShieldAlertIcon },
  { href: routes.admin.cmsPages, label: "CMS", icon: FileTextIcon },
  { href: routes.admin.rbac, label: "RBAC", icon: KeyRoundIcon },
  { href: routes.admin.settings, label: "Settings", icon: SettingsIcon },
  { href: routes.admin.auditLogs, label: "Audit logs", icon: ScrollTextIcon },
  { href: routes.admin.activityLogs, label: "Activity logs", icon: ScrollTextIcon },
  { href: routes.admin.health, label: "System health", icon: ActivityIcon },
]

export function AdminSidebar() {
  return <SidebarNav items={items} ariaLabel="Admin console" />
}
