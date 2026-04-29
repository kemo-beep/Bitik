import { env } from "@/lib/env"
import { bitikFetch } from "@/lib/api/bitik-fetch"
import { parseEnvelope } from "@/lib/api/envelope"
import { listSellerReviews } from "@/lib/api/public"
import { asArray, asString } from "@/lib/safe"

type AnyRecord = Record<string, unknown>
type SellerList<T> = { items: T[] } & AnyRecord

export type SellerBankAccountInput = {
  bank_name: string
  account_name: string
  account_number: string
  country?: string
  currency?: string
  is_default?: boolean
}

async function get<T = AnyRecord>(path: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, { method: "GET" })
  const { data } = await parseEnvelope<T>(res)
  return (data ?? ({} as T)) as T
}

async function post<T = AnyRecord>(path: string, body: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const { data } = await parseEnvelope<T>(res)
  return (data ?? ({} as T)) as T
}

async function patch<T = AnyRecord>(path: string, body: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const { data } = await parseEnvelope<T>(res)
  return (data ?? ({} as T)) as T
}

async function del(path: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, { method: "DELETE" })
  if (!res.ok && res.status !== 204) {
    const { data } = await parseEnvelope<AnyRecord>(res)
    return data ?? {}
  }
  return {}
}

export const sellerApi = {
  apply: (input: AnyRecord) => post("/seller/apply", input),
  getApplication: () => get("/seller/application"),
  patchApplication: (input: AnyRecord) => patch("/seller/application", input),
  createDocument: (input: AnyRecord) => post("/seller/documents", input),
  deleteDocument: (documentId: string) => del(`/seller/documents/${documentId}`),

  getMe: () => get("/seller/me"),
  patchMe: (input: AnyRecord) => patch("/seller/me", input),
  patchSettings: (input: AnyRecord) => patch("/seller/me/settings", input),
  getShippingSettings: async () => {
    const me = await get<AnyRecord>("/seller/me")
    const settings = (me.settings as AnyRecord | undefined) ?? {}
    return (settings.shipping as AnyRecord | undefined) ?? {}
  },
  patchShippingSettings: async (input: AnyRecord) => {
    const me = await get<AnyRecord>("/seller/me")
    const currentSettings = (me.settings as AnyRecord | undefined) ?? {}
    return patch("/seller/me/settings", {
      ...currentSettings,
      shipping: { ...((currentSettings.shipping as AnyRecord | undefined) ?? {}), ...input },
    })
  },
  patchMedia: (input: AnyRecord) => patch("/seller/me/media", input),
  getDashboard: () => get("/seller/dashboard"),

  listProducts: (query = "") => get(`/seller/products${query ? `?${query}` : ""}`),
  createProduct: (input: AnyRecord) => post("/seller/products", input),
  getProduct: (productId: string) => get(`/seller/products/${productId}`),
  patchProduct: (productId: string, input: AnyRecord) => patch(`/seller/products/${productId}`, input),
  deleteProduct: (productId: string) => del(`/seller/products/${productId}`),
  publishProduct: (productId: string) => post(`/seller/products/${productId}/publish`),
  unpublishProduct: (productId: string) => post(`/seller/products/${productId}/unpublish`),
  duplicateProduct: (productId: string) => post(`/seller/products/${productId}/duplicate`),
  createProductImage: (productId: string, input: AnyRecord) => post(`/seller/products/${productId}/images`, input),
  listProductImages: (productId: string) => get(`/seller/products/${productId}/images`),
  patchProductImage: (productId: string, imageId: string, input: AnyRecord) =>
    patch(`/seller/products/${productId}/images/${imageId}`, input),
  deleteProductImage: (productId: string, imageId: string) => del(`/seller/products/${productId}/images/${imageId}`),
  listVariants: (productId: string) => get(`/seller/products/${productId}/variants`),
  createVariant: (productId: string, input: AnyRecord) => post(`/seller/products/${productId}/variants`, input),
  patchVariant: (variantId: string, input: AnyRecord) => patch(`/seller/variants/${variantId}`, input),
  deleteVariant: (variantId: string) => del(`/seller/variants/${variantId}`),
  listOptions: (productId: string) => get(`/seller/products/${productId}/options`),
  createOption: (productId: string, input: AnyRecord) => post(`/seller/products/${productId}/options`, input),
  createOptionValue: (optionId: string, input: AnyRecord) => post(`/seller/options/${optionId}/values`, input),
  deleteOption: (optionId: string) => del(`/seller/options/${optionId}`),

  listInventory: (query = "") => get(`/seller/inventory${query ? `?${query}` : ""}`),
  listLowStock: () => get("/seller/inventory/low-stock"),
  getInventory: (inventoryId: string) => get(`/seller/inventory/${inventoryId}`),
  patchInventory: (inventoryId: string, input: AnyRecord) => patch(`/seller/inventory/${inventoryId}`, input),
  adjustInventory: (inventoryId: string, input: AnyRecord) => post(`/seller/inventory/${inventoryId}/adjust`, input),
  inventoryMovements: (inventoryId: string, query = "") =>
    get(`/seller/inventory/${inventoryId}/movements${query ? `?${query}` : ""}`),
  bulkUpdateInventory: (input: AnyRecord) => post("/seller/inventory/bulk-update", input),

  listOrders: (query = "") => get(`/seller/orders${query ? `?${query}` : ""}`),
  getOrder: (orderId: string) => get(`/seller/orders/${orderId}`),
  listOrderItems: (orderId: string) => get(`/seller/orders/${orderId}/items`),
  acceptOrder: (orderId: string) => post(`/seller/orders/${orderId}/accept`),
  rejectOrder: (orderId: string) => post(`/seller/orders/${orderId}/reject`),
  packOrder: (orderId: string) => post(`/seller/orders/${orderId}/pack`),
  shipOrder: (orderId: string, input: AnyRecord) => post(`/seller/orders/${orderId}/ship`, input),
  cancelOrder: (orderId: string) => post(`/seller/orders/${orderId}/cancel`),
  refundOrder: (orderId: string) => post(`/seller/orders/${orderId}/refund`),
  approveReturn: (orderId: string) => post(`/seller/orders/${orderId}/return/approve`),
  rejectReturn: (orderId: string) => post(`/seller/orders/${orderId}/return/reject`),
  markReturnReceived: (orderId: string) => post(`/seller/orders/${orderId}/return/received`),

  listOrderShipments: (orderId: string) => get(`/seller/orders/${orderId}/shipments`),
  patchShipment: (shipmentId: string, input: AnyRecord) => patch(`/seller/shipments/${shipmentId}`, input),
  markPacked: (shipmentId: string) => post(`/seller/shipments/${shipmentId}/mark-packed`),
  markShipmentShipped: (shipmentId: string) => post(`/seller/shipments/${shipmentId}/mark-shipped`),
  markInTransit: (shipmentId: string) => post(`/seller/shipments/${shipmentId}/mark-in-transit`),
  markDelivered: (shipmentId: string) => post(`/seller/shipments/${shipmentId}/mark-delivered`),
  createShipmentLabel: (shipmentId: string, input: AnyRecord = {}) => post(`/seller/shipments/${shipmentId}/labels`, input),

  listVouchers: () => get("/seller/vouchers"),
  createVoucher: (input: AnyRecord) => post("/seller/vouchers", input),
  patchVoucher: (voucherId: string, input: AnyRecord) => patch(`/seller/vouchers/${voucherId}`, input),
  deleteVoucher: (voucherId: string) => del(`/seller/vouchers/${voucherId}`),

  listReviews: async () => {
    const me = await get<AnyRecord>("/seller/me")
    const sellerId = asString(me.id)
    if (!sellerId) return { items: [] } as SellerList<AnyRecord>
    const items = asArray(await listSellerReviews(sellerId)) ?? []
    return { items } as SellerList<AnyRecord>
  },
  replyReview: (reviewId: string, body: string) => post(`/seller/reviews/${reviewId}/reply`, { body }),
  reportReview: (reviewId: string, input: AnyRecord) => post(`/seller/reviews/${reviewId}/report`, input),

  capturePOD: (paymentId: string, input: AnyRecord = {}) => post(`/seller/payments/${paymentId}/pod/capture`, input),
  getWallet: () => get("/seller/wallet"),
  listWalletTransactions: (query = "") => get(`/seller/wallet/transactions${query ? `?${query}` : ""}`),
  listBankAccounts: () => get<SellerList<AnyRecord>>("/seller/bank-accounts"),
  createBankAccount: (input: SellerBankAccountInput) => post("/seller/bank-accounts", input),
  patchBankAccount: (bankAccountId: string, input: Partial<SellerBankAccountInput>) =>
    patch(`/seller/bank-accounts/${bankAccountId}`, input as AnyRecord),
  deleteBankAccount: (bankAccountId: string) => del(`/seller/bank-accounts/${bankAccountId}`),
  setDefaultBankAccount: (bankAccountId: string) => post(`/seller/bank-accounts/${bankAccountId}/set-default`),
  requestPayout: (input: AnyRecord) => post("/seller/payouts/request", input),
  listPayouts: (query = "") => get(`/seller/payouts${query ? `?${query}` : ""}`),

  listChatConversations: (query = "") => get(`/chat/conversations${query ? `?${query}` : ""}`),
  createChatConversation: (input: AnyRecord) => post("/chat/conversations", input),
  listChatMessages: (conversationId: string, query = "") =>
    get(`/chat/conversations/${conversationId}/messages${query ? `?${query}` : ""}`),
  sendChatMessage: (conversationId: string, input: AnyRecord) => post(`/chat/conversations/${conversationId}/messages`, input),
  markChatRead: async (conversationId: string) => {
    const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations/${conversationId}/read`, { method: "PATCH" })
    if (!res.ok && res.status !== 204) await parseEnvelope<AnyRecord>(res)
    return {}
  },
  deleteChatMessage: async (conversationId: string, messageId: string) => {
    const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations/${conversationId}/messages/${messageId}`, { method: "DELETE" })
    if (!res.ok && res.status !== 204) await parseEnvelope<AnyRecord>(res)
    return {}
  },
  deleteChatConversation: async (conversationId: string) => {
    const res = await bitikFetch(`${env.apiBaseUrl}/chat/conversations/${conversationId}`, { method: "DELETE" })
    if (!res.ok && res.status !== 204) await parseEnvelope<AnyRecord>(res)
    return {}
  },
  listNotifications: (query = "") => get(`/seller/notifications${query ? `?${query}` : ""}`),
  getNotificationsUnread: () => get("/seller/notifications/unread-count"),
  markNotificationRead: (notificationId: string) => patch(`/seller/notifications/${notificationId}/read`),
  markAllNotificationsRead: async () => {
    const res = await bitikFetch(`${env.apiBaseUrl}/seller/notifications/read-all`, { method: "PATCH" })
    if (!res.ok && res.status !== 204) await parseEnvelope<AnyRecord>(res)
    return {}
  },
  deleteNotification: async (notificationId: string) => {
    const res = await bitikFetch(`${env.apiBaseUrl}/seller/notifications/${notificationId}`, { method: "DELETE" })
    if (!res.ok && res.status !== 204) await parseEnvelope<AnyRecord>(res)
    return {}
  },
  getNotificationPreferences: () => get("/seller/notifications/preferences"),
  putNotificationPreferences: async (input: AnyRecord) => {
    const res = await bitikFetch(`${env.apiBaseUrl}/seller/notifications/preferences`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    })
    const { data } = await parseEnvelope<AnyRecord>(res)
    return data ?? {}
  },
}

