export type StatusTone =
  | "neutral"
  | "info"
  | "success"
  | "warning"
  | "danger"
  | "pending"

export type StatusMeta = {
  label: string
  tone: StatusTone
  description?: string
}

export const ORDER_STATUS = [
  "pending",
  "awaiting_payment",
  "paid",
  "processing",
  "packed",
  "shipped",
  "out_for_delivery",
  "delivered",
  "completed",
  "cancelled",
  "refund_requested",
  "refunded",
  "returned",
  "disputed",
] as const
export type OrderStatus = (typeof ORDER_STATUS)[number]

export const ORDER_STATUS_META: Record<OrderStatus, StatusMeta> = {
  pending: { label: "Pending", tone: "pending" },
  awaiting_payment: { label: "Awaiting payment", tone: "warning" },
  paid: { label: "Paid", tone: "info" },
  processing: { label: "Processing", tone: "info" },
  packed: { label: "Packed", tone: "info" },
  shipped: { label: "Shipped", tone: "info" },
  out_for_delivery: { label: "Out for delivery", tone: "info" },
  delivered: { label: "Delivered", tone: "success" },
  completed: { label: "Completed", tone: "success" },
  cancelled: { label: "Cancelled", tone: "neutral" },
  refund_requested: { label: "Refund requested", tone: "warning" },
  refunded: { label: "Refunded", tone: "neutral" },
  returned: { label: "Returned", tone: "neutral" },
  disputed: { label: "Disputed", tone: "danger" },
}

export const PAYMENT_STATUS = [
  "pending",
  "awaiting_manual_confirmation",
  "authorized",
  "paid",
  "payable_on_delivery",
  "captured_on_delivery",
  "failed",
  "rejected",
  "expired",
  "refunded",
  "partially_refunded",
  "cancelled",
] as const
export type PaymentStatus = (typeof PAYMENT_STATUS)[number]

export const PAYMENT_STATUS_META: Record<PaymentStatus, StatusMeta> = {
  pending: { label: "Pending", tone: "pending" },
  awaiting_manual_confirmation: { label: "Awaiting Wave confirmation", tone: "warning" },
  authorized: { label: "Authorized", tone: "info" },
  paid: { label: "Paid", tone: "success" },
  payable_on_delivery: { label: "Pay on delivery", tone: "info" },
  captured_on_delivery: { label: "Captured on delivery", tone: "success" },
  failed: { label: "Failed", tone: "danger" },
  rejected: { label: "Rejected", tone: "danger" },
  expired: { label: "Expired", tone: "neutral" },
  refunded: { label: "Refunded", tone: "neutral" },
  partially_refunded: { label: "Partially refunded", tone: "warning" },
  cancelled: { label: "Cancelled", tone: "neutral" },
}

export const SHIPMENT_STATUS = [
  "pending",
  "label_created",
  "ready_to_ship",
  "shipped",
  "in_transit",
  "out_for_delivery",
  "delivered",
  "delivery_failed",
  "returned",
  "cancelled",
] as const
export type ShipmentStatus = (typeof SHIPMENT_STATUS)[number]

export const SHIPMENT_STATUS_META: Record<ShipmentStatus, StatusMeta> = {
  pending: { label: "Pending", tone: "pending" },
  label_created: { label: "Label created", tone: "info" },
  ready_to_ship: { label: "Ready to ship", tone: "info" },
  shipped: { label: "Shipped", tone: "info" },
  in_transit: { label: "In transit", tone: "info" },
  out_for_delivery: { label: "Out for delivery", tone: "info" },
  delivered: { label: "Delivered", tone: "success" },
  delivery_failed: { label: "Delivery failed", tone: "danger" },
  returned: { label: "Returned", tone: "neutral" },
  cancelled: { label: "Cancelled", tone: "neutral" },
}

export const SELLER_STATUS = [
  "draft",
  "submitted",
  "under_review",
  "approved",
  "rejected",
  "active",
  "suspended",
  "closed",
] as const
export type SellerStatus = (typeof SELLER_STATUS)[number]

export const SELLER_STATUS_META: Record<SellerStatus, StatusMeta> = {
  draft: { label: "Draft", tone: "neutral" },
  submitted: { label: "Submitted", tone: "pending" },
  under_review: { label: "Under review", tone: "warning" },
  approved: { label: "Approved", tone: "success" },
  rejected: { label: "Rejected", tone: "danger" },
  active: { label: "Active", tone: "success" },
  suspended: { label: "Suspended", tone: "danger" },
  closed: { label: "Closed", tone: "neutral" },
}

export const PRODUCT_STATUS = [
  "draft",
  "pending_review",
  "active",
  "unpublished",
  "rejected",
  "banned",
  "archived",
] as const
export type ProductStatus = (typeof PRODUCT_STATUS)[number]

export const PRODUCT_STATUS_META: Record<ProductStatus, StatusMeta> = {
  draft: { label: "Draft", tone: "neutral" },
  pending_review: { label: "Pending review", tone: "pending" },
  active: { label: "Active", tone: "success" },
  unpublished: { label: "Unpublished", tone: "neutral" },
  rejected: { label: "Rejected", tone: "danger" },
  banned: { label: "Banned", tone: "danger" },
  archived: { label: "Archived", tone: "neutral" },
}

export const REFUND_STATUS = [
  "requested",
  "under_review",
  "approved",
  "rejected",
  "processing",
  "completed",
  "cancelled",
] as const
export type RefundStatus = (typeof REFUND_STATUS)[number]

export const REFUND_STATUS_META: Record<RefundStatus, StatusMeta> = {
  requested: { label: "Requested", tone: "pending" },
  under_review: { label: "Under review", tone: "warning" },
  approved: { label: "Approved", tone: "success" },
  rejected: { label: "Rejected", tone: "danger" },
  processing: { label: "Processing", tone: "info" },
  completed: { label: "Completed", tone: "success" },
  cancelled: { label: "Cancelled", tone: "neutral" },
}

export const RETURN_STATUS = [
  "requested",
  "approved",
  "rejected",
  "in_transit",
  "received",
  "refunded",
  "cancelled",
] as const
export type ReturnStatus = (typeof RETURN_STATUS)[number]

export const RETURN_STATUS_META: Record<ReturnStatus, StatusMeta> = {
  requested: { label: "Requested", tone: "pending" },
  approved: { label: "Approved", tone: "success" },
  rejected: { label: "Rejected", tone: "danger" },
  in_transit: { label: "In transit", tone: "info" },
  received: { label: "Received", tone: "info" },
  refunded: { label: "Refunded", tone: "success" },
  cancelled: { label: "Cancelled", tone: "neutral" },
}

export const STATUS_TONE_CLASS: Record<StatusTone, string> = {
  neutral: "bg-muted text-muted-foreground border-border",
  info: "bg-info/10 text-info border-info/20",
  success: "bg-success/10 text-success border-success/20",
  warning: "bg-warning/15 text-warning-foreground border-warning/30",
  danger: "bg-destructive/10 text-destructive border-destructive/20",
  pending: "bg-pending/10 text-pending-foreground border-pending/30 dark:text-pending",
}

export type AnyStatusKind =
  | "order"
  | "payment"
  | "shipment"
  | "seller"
  | "product"
  | "refund"
  | "return"

export function statusMeta(kind: AnyStatusKind, value: string): StatusMeta {
  switch (kind) {
    case "order":
      return ORDER_STATUS_META[value as OrderStatus] ?? fallback(value)
    case "payment":
      return PAYMENT_STATUS_META[value as PaymentStatus] ?? fallback(value)
    case "shipment":
      return SHIPMENT_STATUS_META[value as ShipmentStatus] ?? fallback(value)
    case "seller":
      return SELLER_STATUS_META[value as SellerStatus] ?? fallback(value)
    case "product":
      return PRODUCT_STATUS_META[value as ProductStatus] ?? fallback(value)
    case "refund":
      return REFUND_STATUS_META[value as RefundStatus] ?? fallback(value)
    case "return":
      return RETURN_STATUS_META[value as ReturnStatus] ?? fallback(value)
  }
}

function fallback(value: string): StatusMeta {
  return { label: humanize(value), tone: "neutral" }
}

export function humanize(value: string): string {
  return value
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
    .trim()
    .replace(/\b\w/g, (c) => c.toUpperCase())
}
