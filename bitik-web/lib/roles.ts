export const USER_ROLES = [
  "guest",
  "buyer",
  "seller_pending",
  "seller_active",
  "seller_suspended",
  "admin",
  "staff",
] as const
export type UserRole = (typeof USER_ROLES)[number]

export const ROLE_LABEL: Record<UserRole, string> = {
  guest: "Guest",
  buyer: "Buyer",
  seller_pending: "Seller (pending)",
  seller_active: "Seller",
  seller_suspended: "Seller (suspended)",
  admin: "Admin",
  staff: "Staff",
}

export const PERMISSION_KEYS = [
  "buyer.cart.manage",
  "buyer.checkout.place",
  "buyer.orders.view",
  "buyer.reviews.write",
  "seller.application.submit",
  "seller.shop.manage",
  "seller.products.manage",
  "seller.inventory.manage",
  "seller.orders.fulfill",
  "seller.shipping.manage",
  "seller.payouts.request",
  "seller.promotions.manage",
  "admin.users.manage",
  "admin.sellers.approve",
  "admin.products.moderate",
  "admin.orders.manage",
  "admin.payments.approve_wave",
  "admin.refunds.process",
  "admin.cms.manage",
  "admin.rbac.manage",
  "admin.settings.manage",
  "admin.audit.read",
  "admin.system.health",
] as const
export type PermissionKey = (typeof PERMISSION_KEYS)[number]

export const AREA_ACCESS: Record<string, UserRole[]> = {
  storefront: ["guest", "buyer", "seller_pending", "seller_active", "seller_suspended", "admin", "staff"],
  account: ["buyer", "seller_pending", "seller_active", "seller_suspended", "admin", "staff"],
  seller_apply: ["buyer", "seller_pending"],
  seller_center: ["seller_active", "seller_suspended", "admin", "staff"],
  admin_console: ["admin", "staff"],
}

export type SessionUser = {
  id: string
  email?: string
  name?: string
  role: UserRole
  permissions?: PermissionKey[]
  sellerStatus?: import("./status").SellerStatus
}

export function roleFromBackendRoles(roles?: string[] | null): UserRole {
  const set = new Set((roles ?? []).map((r) => r.toLowerCase()))
  if (set.has("admin")) return "admin"
  if (set.has("seller")) return "seller_active"
  if (set.has("buyer")) return "buyer"
  return "guest"
}

export function canAccess(area: keyof typeof AREA_ACCESS, role: UserRole): boolean {
  return AREA_ACCESS[area]?.includes(role) ?? false
}

export function hasPermission(user: SessionUser | null | undefined, key: PermissionKey): boolean {
  if (!user) return false
  if (user.role === "admin") return true
  return user.permissions?.includes(key) ?? false
}

export function isSeller(role: UserRole): boolean {
  return role === "seller_active" || role === "seller_suspended" || role === "seller_pending"
}

export function isStaff(role: UserRole): boolean {
  return role === "admin" || role === "staff"
}
