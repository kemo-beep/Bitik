import { env } from "@/lib/env"
import { bitikFetch } from "@/lib/api/bitik-fetch"
import { parseEnvelope } from "@/lib/api/envelope"

type AnyRecord = Record<string, unknown>
export type AdminListResponse<T> = { items: T[]; pagination?: Record<string, unknown> }
export type AdminCategory = { id: string; name: string; slug: string; sort_order: number; is_active: boolean }
export type AdminBrand = { id: string; name: string; slug: string; is_active: boolean }
export type AdminProduct = { id: string; name: string; slug: string; moderation_status?: string; status?: string }
export type AdminShipment = { id: string; order_id: string; status: string; tracking_number?: string | null }
export type AdminWebhookEvent = {
  id: string
  provider: string
  event_id: string
  event_type?: string | null
  processed: boolean
}

async function get(path: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, { method: "GET" })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}
async function getTyped<T>(path: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, { method: "GET" })
  const { data } = await parseEnvelope<T>(res)
  return data as T
}

async function post(path: string, body: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

async function patch(path: string, body: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

async function put(path: string, body: AnyRecord = {}) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  })
  const { data } = await parseEnvelope<AnyRecord>(res)
  return data ?? {}
}

async function del(path: string) {
  const res = await bitikFetch(`${env.apiBaseUrl}${path}`, { method: "DELETE" })
  if (!res.ok && res.status !== 204) {
    const { data } = await parseEnvelope<AnyRecord>(res)
    return data ?? {}
  }
  return {}
}

export const adminApi = {
  listCategories: () => getTyped<AdminListResponse<AdminCategory>>("/admin/categories"),
  createCategory: (input: AnyRecord) => post("/admin/categories", input),
  patchCategory: (categoryId: string, input: AnyRecord) => patch(`/admin/categories/${categoryId}`, input),
  deleteCategory: (categoryId: string) => del(`/admin/categories/${categoryId}`),
  reorderCategories: (items: Array<{ id: string; sort_order: number }>) => post("/admin/categories/reorder", { items }),
  listBrands: () => getTyped<AdminListResponse<AdminBrand>>("/admin/brands"),
  createBrand: (input: AnyRecord) => post("/admin/brands", input),
  patchBrand: (brandId: string, input: AnyRecord) => patch(`/admin/brands/${brandId}`, input),
  deleteBrand: (brandId: string) => del(`/admin/brands/${brandId}`),
  listProducts: (query = "") => getTyped<AdminListResponse<AdminProduct>>(`/admin/products${query ? `?${query}` : ""}`),
  getProduct: (productId: string) => getTyped<AdminProduct>(`/admin/products/${productId}`),

  dashboardOverview: () => get("/admin/dashboard/overview"),
  dashboardEventChart: () => get("/admin/dashboard/charts/events"),
  health: () => get("/admin/health"),

  listUsers: (query = "") => get(`/admin/users${query ? `?${query}` : ""}`),
  getUser: (userId: string) => get(`/admin/users/${userId}`),
  patchUserStatus: (userId: string, input: AnyRecord) => patch(`/admin/users/${userId}/status`, input),
  getUserRoles: (userId: string) => get(`/admin/users/${userId}/roles`),
  replaceUserRoles: (userId: string, input: AnyRecord) => put(`/admin/users/${userId}/roles`, input),

  listSellerApplications: (query = "") => get(`/admin/seller-applications${query ? `?${query}` : ""}`),
  reviewSellerApplication: (applicationId: string, input: AnyRecord) =>
    post(`/admin/seller-applications/${applicationId}/review`, input),
  patchSellerStatus: (sellerId: string, input: AnyRecord) => patch(`/admin/sellers/${sellerId}/status`, input),
  patchProductModeration: (productId: string, input: AnyRecord) => patch(`/admin/products/${productId}/moderation`, input),

  listOrders: (query = "") => get(`/admin/orders${query ? `?${query}` : ""}`),
  getOrder: (orderId: string) => get(`/admin/orders/${orderId}`),
  patchOrder: (orderId: string, input: AnyRecord) => patch(`/admin/orders/${orderId}`, input),
  cancelOrder: (orderId: string, input: AnyRecord = {}) => post(`/admin/orders/${orderId}/cancel`, input),
  refundOrder: (orderId: string, input: AnyRecord = {}) => post(`/admin/orders/${orderId}/refund`, input),
  orderStatusHistory: (orderId: string) => get(`/admin/orders/${orderId}/status-history`),
  orderPayments: (orderId: string) => get(`/admin/orders/${orderId}/payments`),
  orderShipments: (orderId: string) => get(`/admin/orders/${orderId}/shipments`),

  listPendingWavePayments: () => get("/admin/payments/wave/pending"),
  approveWavePayment: (paymentId: string, input: AnyRecord = {}) => post(`/admin/payments/${paymentId}/wave/approve`, input),
  rejectWavePayment: (paymentId: string, input: AnyRecord = {}) => post(`/admin/payments/${paymentId}/wave/reject`, input),
  capturePodPayment: (paymentId: string, input: AnyRecord = {}) => post(`/admin/payments/${paymentId}/pod/capture`, input),
  reviewRefund: (refundId: string, input: AnyRecord) => post(`/admin/refunds/${refundId}/review`, input),
  processRefund: (refundId: string, input: AnyRecord = {}) => post(`/admin/refunds/${refundId}/process`, input),
  patchPayoutStatus: (payoutId: string, input: AnyRecord) => post(`/admin/payouts/${payoutId}/status`, input),

  listShippingProviders: () => get("/admin/shipping/providers"),
  createShippingProvider: (input: AnyRecord) => post("/admin/shipping/providers", input),
  patchShippingProvider: (providerId: string, input: AnyRecord) => patch(`/admin/shipping/providers/${providerId}`, input),
  listShipments: (query = "") => getTyped<AdminListResponse<AdminShipment>>(`/admin/shipping/shipments${query ? `?${query}` : ""}`),
  getShipment: (shipmentId: string) => getTyped<AdminShipment>(`/admin/shipping/shipments/${shipmentId}`),
  shipmentTracking: (shipmentId: string) => getTyped<AdminListResponse<AnyRecord>>(`/admin/shipping/shipments/${shipmentId}/tracking`),
  patchShipmentStatus: (shipmentId: string, input: AnyRecord) => post(`/admin/shipping/shipments/${shipmentId}/status`, input),

  listVouchers: () => get("/admin/vouchers"),
  createVoucher: (input: AnyRecord) => post("/admin/vouchers", input),
  patchVoucher: (voucherId: string, input: AnyRecord) => patch(`/admin/vouchers/${voucherId}`, input),
  deleteVoucher: (voucherId: string) => del(`/admin/vouchers/${voucherId}`),

  hideReview: (reviewId: string, hidden: boolean) => patch(`/admin/reviews/${reviewId}/hide`, { hidden }),
  deleteReview: (reviewId: string) => del(`/admin/reviews/${reviewId}`),
  listReviewReports: () => get("/admin/reviews/reports"),
  resolveReviewReport: (reportId: string, input: AnyRecord) => patch(`/admin/reviews/reports/${reportId}/resolve`, input),

  listModerationReports: (query = "") => get(`/admin/moderation/reports${query ? `?${query}` : ""}`),
  getModerationReport: (reportId: string) => get(`/admin/moderation/reports/${reportId}`),
  patchModerationReportStatus: (reportId: string, input: AnyRecord) => patch(`/admin/moderation/reports/${reportId}/status`, input),
  listModerationCases: (query = "") => get(`/admin/moderation/cases${query ? `?${query}` : ""}`),
  getModerationCase: (caseId: string) => get(`/admin/moderation/cases/${caseId}`),
  createModerationCase: (input: AnyRecord) => post("/admin/moderation/cases", input),
  patchModerationCase: (caseId: string, input: AnyRecord) => patch(`/admin/moderation/cases/${caseId}`, input),

  listCmsPages: () => get("/admin/cms/pages"),
  createCmsPage: (input: AnyRecord) => post("/admin/cms/pages", input),
  patchCmsPage: (pageId: string, input: AnyRecord) => patch(`/admin/cms/pages/${pageId}`, input),
  deleteCmsPage: (pageId: string) => del(`/admin/cms/pages/${pageId}`),
  listCmsBanners: () => get("/admin/cms/banners"),
  listCmsFaqs: () => get("/admin/cms/faqs"),
  listCmsAnnouncements: () => get("/admin/cms/announcements"),

  listRoles: () => get("/admin/rbac/roles"),
  createRole: (input: AnyRecord) => post("/admin/rbac/roles", input),
  patchRole: (roleId: string, input: AnyRecord) => patch(`/admin/rbac/roles/${roleId}`, input),
  deleteRole: (roleId: string) => del(`/admin/rbac/roles/${roleId}`),
  listPermissions: () => get("/admin/rbac/permissions"),
  listRolePermissions: (roleId: string) => get(`/admin/rbac/roles/${roleId}/permissions`),

  listPlatformSettings: () => get("/admin/settings/platform"),
  putPlatformSetting: (key: string, input: AnyRecord) => put(`/admin/settings/platform/${key}`, input),
  listFeatureFlags: () => get("/admin/settings/feature-flags"),
  putFeatureFlag: (key: string, input: AnyRecord) => put(`/admin/settings/feature-flags/${key}`, input),
  listAuditLogs: (query = "") => get(`/admin/logs/audit${query ? `?${query}` : ""}`),
  listActivityLogs: (query = "") => get(`/admin/logs/activity${query ? `?${query}` : ""}`),
  listWebhookEvents: (query = "") => getTyped<AdminListResponse<AdminWebhookEvent>>(`/admin/payments/webhooks${query ? `?${query}` : ""}`),
  getWebhookEvent: (eventId: string) => getTyped<AnyRecord>(`/admin/payments/webhooks/${eventId}`),
  reprocessWebhookEvent: (eventId: string) => post(`/admin/payments/webhooks/${eventId}/reprocess`),
  listNotifications: (query = "") => get(`/admin/notifications${query ? `?${query}` : ""}`),
  getNotificationsUnread: () => get("/admin/notifications/unread-count"),
  markNotificationRead: (notificationId: string) => patch(`/admin/notifications/${notificationId}/read`),
  markAllNotificationsRead: async () => {
    const res = await bitikFetch(`${env.apiBaseUrl}/admin/notifications/read-all`, { method: "PATCH" })
    if (!res.ok && res.status !== 204) await parseEnvelope<AnyRecord>(res)
    return {}
  },
  deleteNotification: (notificationId: string) => del(`/admin/notifications/${notificationId}`),
  getNotificationPreferences: () => get("/admin/notifications/preferences"),
  putNotificationPreferences: (input: AnyRecord) => put("/admin/notifications/preferences", input),
}

