export const analyticsEvents = {
  productView: "product.view",
  searchSubmit: "search.submit",
  searchClick: "search.click",
  addToCart: "cart.add",
  checkoutStart: "checkout.start",
  placeOrder: "checkout.place_order",
  paymentMethodSelected: "payment.method_selected",
  sellerProductCreate: "seller.product_create",
  sellerProductPublish: "seller.product_publish",
  adminModerationAction: "admin.moderation_action",
} as const

export type AnalyticsEventName = (typeof analyticsEvents)[keyof typeof analyticsEvents]

