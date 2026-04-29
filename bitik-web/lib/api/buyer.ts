import { env } from "@/lib/env"
import { bitikFetch } from "@/lib/api/bitik-fetch"
import { parseEnvelope } from "@/lib/api/envelope"

type AnyRecord = Record<string, unknown>

const WISHLIST_KEY = "bitik.wishlist.product_ids.v1"
const REVIEWS_KEY = "bitik.reviews.local.v1"
const NOTIFICATIONS_KEY = "bitik.notifications.local.v1"

function safeJson<T>(raw: string | null, fallback: T): T {
  if (!raw) return fallback
  try {
    return JSON.parse(raw) as T
  } catch {
    return fallback
  }
}

export async function getCart() {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function clearCart() {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart`, { method: "DELETE" })
  if (!res.ok && res.status !== 204) await parseEnvelope<unknown>(res)
}

export async function addCartItem(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart/items`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function updateCartItem(cartItemId: string, input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart/items/${cartItemId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function deleteCartItem(cartItemId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart/items/${cartItemId}`, { method: "DELETE" })
  if (!res.ok && res.status !== 204) await parseEnvelope<unknown>(res)
}

export async function mergeGuestCart(input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart/merge`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function selectCartItems(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart/select-items`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function applyCartVoucher(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart/apply-voucher`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function removeCartVoucher(voucherId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/cart/voucher/${voucherId}`, { method: "DELETE" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function createCheckoutSession(input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/checkout/sessions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function getCheckoutSession(checkoutSessionId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function patchCheckoutAddress(checkoutSessionId: string, input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}/address`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function patchCheckoutShipping(checkoutSessionId: string, input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}/shipping`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function patchCheckoutPaymentMethod(checkoutSessionId: string, input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}/payment-method`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function applyCheckoutVoucher(checkoutSessionId: string, input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}/apply-voucher`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function removeCheckoutVoucher(checkoutSessionId: string, voucherId: string) {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}/voucher/${voucherId}`,
    { method: "DELETE" }
  )
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function validateCheckoutSession(checkoutSessionId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}/validate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function placeOrderFromCheckoutSession(
  checkoutSessionId: string,
  input: AnyRecord = {},
  idempotencyKey?: string
) {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/buyer/checkout/sessions/${checkoutSessionId}/place-order`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    },
    { idempotencyKey }
  )
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function listBuyerOrders(params?: { status?: string }) {
  const url = new URL(`${env.apiBaseUrl}/buyer/orders`)
  if (params?.status) url.searchParams.set("status", params.status)
  const res = await bitikFetch(url.toString(), { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function getBuyerOrder(orderId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function listBuyerOrderItems(orderId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/items`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function listBuyerOrderStatusHistory(orderId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/status-history`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function cancelBuyerOrder(orderId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/cancel`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function confirmBuyerOrderReceived(orderId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/confirm-received`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function requestBuyerOrderRefund(orderId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/request-refund`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function requestBuyerOrderReturn(orderId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/request-return`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function disputeBuyerOrder(orderId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/dispute`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function getBuyerOrderTracking(orderId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/tracking`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function getBuyerOrderInvoice(orderId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/orders/${orderId}/invoice`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function listBuyerPaymentMethods() {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payment-methods`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function createBuyerPaymentMethod(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payment-methods`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function setDefaultBuyerPaymentMethod(methodId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payment-methods/${methodId}/set-default`, { method: "POST" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function deleteBuyerPaymentMethod(methodId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payment-methods/${methodId}`, { method: "DELETE" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function createBuyerPaymentIntent(input: AnyRecord, idempotencyKey?: string) {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/buyer/payments/create-intent`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    },
    { idempotencyKey }
  )
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function confirmBuyerPayment(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payments/confirm`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function getBuyerPayment(paymentId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payments/${paymentId}`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function retryBuyerPayment(paymentId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payments/${paymentId}/retry`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function cancelBuyerPayment(paymentId: string, input: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/payments/${paymentId}/cancel`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function listChatConversations(query = "") {
  const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations${query ? `?${query}` : ""}`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function createChatConversation(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function listChatMessages(conversationId: string, query = "") {
  const res = await bitikFetch(
    `${env.apiBaseUrl}/chat/conversations/${conversationId}/messages${query ? `?${query}` : ""}`,
    { method: "GET" }
  )
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function sendChatMessage(conversationId: string, input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations/${conversationId}/messages`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function markChatConversationRead(conversationId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations/${conversationId}/read`, { method: "PATCH" })
  if (!res.ok && res.status !== 204) await parseEnvelope<unknown>(res)
}

export async function deleteChatMessage(conversationId: string, messageId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations/${conversationId}/messages/${messageId}`, { method: "DELETE" })
  if (!res.ok && res.status !== 204) await parseEnvelope<unknown>(res)
}

export async function deleteChatConversation(conversationId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations/${conversationId}`, { method: "DELETE" })
  if (!res.ok && res.status !== 204) await parseEnvelope<unknown>(res)
}

export async function createMediaPresignedUrl(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/media/upload/presigned-url`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function completeMediaPresignedUpload(fileId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/media/upload/presigned-complete`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ file_id: fileId }),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function listBuyerNotifications(query = "") {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/notifications${query ? `?${query}` : ""}`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function getBuyerNotificationsUnreadCount() {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/notifications/unread-count`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function markBuyerNotificationRead(notificationId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/notifications/${notificationId}/read`, { method: "PATCH" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function markAllBuyerNotificationsRead() {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/notifications/read-all`, { method: "PATCH" })
  if (!res.ok && res.status !== 204) await parseEnvelope<unknown>(res)
}

export async function deleteBuyerNotification(notificationId: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/notifications/${notificationId}`, { method: "DELETE" })
  if (!res.ok && res.status !== 204) await parseEnvelope<unknown>(res)
}

export async function getBuyerNotificationPreferences() {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/notifications/preferences`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export async function putBuyerNotificationPreferences(input: AnyRecord) {
  const res = await bitikFetch(`${env.apiBaseUrl}/buyer/notifications/preferences`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

export function listWishlistProductIds() {
  if (typeof window === "undefined") return [] as string[]
  const value = safeJson<string[]>(window.localStorage.getItem(WISHLIST_KEY), [])
  return value.filter(Boolean)
}

export function addWishlistProductId(productId: string) {
  if (typeof window === "undefined") return
  const current = listWishlistProductIds()
  const next = Array.from(new Set([productId, ...current]))
  window.localStorage.setItem(WISHLIST_KEY, JSON.stringify(next))
}

export function removeWishlistProductId(productId: string) {
  if (typeof window === "undefined") return
  const current = listWishlistProductIds()
  window.localStorage.setItem(
    WISHLIST_KEY,
    JSON.stringify(current.filter((id) => id !== productId))
  )
}

export function listLocalBuyerReviews() {
  if (typeof window === "undefined") return [] as AnyRecord[]
  return safeJson<AnyRecord[]>(window.localStorage.getItem(REVIEWS_KEY), [])
}

export function saveLocalBuyerReview(review: AnyRecord) {
  if (typeof window === "undefined") return
  const reviews = listLocalBuyerReviews()
  const id = String(review.id ?? crypto.randomUUID())
  const rest = reviews.filter((r) => String(r.id ?? "") !== id)
  window.localStorage.setItem(REVIEWS_KEY, JSON.stringify([{ ...review, id }, ...rest]))
}

export function deleteLocalBuyerReview(reviewId: string) {
  if (typeof window === "undefined") return
  const reviews = listLocalBuyerReviews()
  window.localStorage.setItem(
    REVIEWS_KEY,
    JSON.stringify(reviews.filter((r) => String(r.id ?? "") !== reviewId))
  )
}

export function listLocalNotifications() {
  if (typeof window === "undefined") return [] as AnyRecord[]
  return safeJson<AnyRecord[]>(window.localStorage.getItem(NOTIFICATIONS_KEY), [])
}

export function seedLocalNotifications() {
  if (typeof window === "undefined") return
  const existing = listLocalNotifications()
  if (existing.length > 0) return
  const now = new Date().toISOString()
  const sample = [
    { id: crypto.randomUUID(), title: "Order update", body: "Your order is being prepared.", read: false, created_at: now },
    { id: crypto.randomUUID(), title: "Voucher reminder", body: "You have a voucher expiring soon.", read: false, created_at: now },
  ]
  window.localStorage.setItem(NOTIFICATIONS_KEY, JSON.stringify(sample))
}

export function markLocalNotificationRead(id: string) {
  if (typeof window === "undefined") return
  const next = listLocalNotifications().map((n) => (String(n.id ?? "") === id ? { ...n, read: true } : n))
  window.localStorage.setItem(NOTIFICATIONS_KEY, JSON.stringify(next))
}

export function markAllLocalNotificationsRead() {
  if (typeof window === "undefined") return
  const next = listLocalNotifications().map((n) => ({ ...n, read: true }))
  window.localStorage.setItem(NOTIFICATIONS_KEY, JSON.stringify(next))
}

export function deleteLocalNotification(id: string) {
  if (typeof window === "undefined") return
  const next = listLocalNotifications().filter((n) => String(n.id ?? "") !== id)
  window.localStorage.setItem(NOTIFICATIONS_KEY, JSON.stringify(next))
}

