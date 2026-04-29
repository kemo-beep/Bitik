Shopee-like marketplace.

Base structure
/api/v1/public
/api/v1/auth
/api/v1/users
/api/v1/buyer
/api/v1/seller
/api/v1/admin
/api/v1/webhooks
/api/v1/internal
/ws
1. Auth APIs
POST   /auth/register
POST   /auth/login
POST   /auth/logout
POST   /auth/refresh-token
POST   /auth/forgot-password
POST   /auth/reset-password
POST   /auth/verify-email
POST   /auth/resend-email-verification
POST   /auth/send-phone-otp
POST   /auth/verify-phone-otp

GET    /auth/oauth/google
GET    /auth/oauth/google/callback
GET    /auth/oauth/apple
GET    /auth/oauth/apple/callback
GET    /auth/oauth/facebook
GET    /auth/oauth/facebook/callback
2. User / Buyer APIs
GET    /users/me
PATCH  /users/me
DELETE /users/me

GET    /users/me/profile
PATCH  /users/me/profile

GET    /users/me/sessions
DELETE /users/me/sessions/{session_id}

GET    /users/me/devices
DELETE /users/me/devices/{device_id}
3. Address APIs
GET    /buyer/addresses
POST   /buyer/addresses
GET    /buyer/addresses/{address_id}
PATCH  /buyer/addresses/{address_id}
DELETE /buyer/addresses/{address_id}
POST   /buyer/addresses/{address_id}/set-default
4. Public Catalog APIs
GET    /public/home
GET    /public/home/sections
GET    /public/banners
GET    /public/categories
GET    /public/categories/{category_id}
GET    /public/categories/{category_id}/products
GET    /public/brands
GET    /public/brands/{brand_id}
GET    /public/brands/{brand_id}/products

GET    /public/products
GET    /public/products/{product_id}
GET    /public/products/slug/{slug}
GET    /public/products/{product_id}/variants
GET    /public/products/{product_id}/reviews
GET    /public/products/{product_id}/related

GET    /public/sellers/{seller_id}
GET    /public/sellers/{seller_id}/products
GET    /public/sellers/{seller_id}/reviews
GET    /public/sellers/slug/{slug}
5. Search APIs
GET    /public/search
GET    /public/search/suggestions
GET    /public/search/trending
GET    /public/search/recent
DELETE /public/search/recent
POST   /public/search/click

Example filters:

?q=shoes
&category_id=
&brand_id=
&min_price=
&max_price=
&rating=
&sort=popular|latest|price_asc|price_desc|top_sales
&page=
&limit=
6. Cart APIs
GET    /buyer/cart
POST   /buyer/cart/items
PATCH  /buyer/cart/items/{cart_item_id}
DELETE /buyer/cart/items/{cart_item_id}
DELETE /buyer/cart
POST   /buyer/cart/merge
POST   /buyer/cart/select-items
POST   /buyer/cart/apply-voucher
DELETE /buyer/cart/voucher/{voucher_id}
7. Wishlist APIs
GET    /buyer/wishlist
POST   /buyer/wishlist/items
DELETE /buyer/wishlist/items/{product_id}
POST   /buyer/wishlist/items/{product_id}/move-to-cart
8. Checkout APIs
POST   /buyer/checkout/sessions
GET    /buyer/checkout/sessions/{checkout_session_id}
PATCH  /buyer/checkout/sessions/{checkout_session_id}/address
PATCH  /buyer/checkout/sessions/{checkout_session_id}/shipping
PATCH  /buyer/checkout/sessions/{checkout_session_id}/payment-method
POST   /buyer/checkout/sessions/{checkout_session_id}/apply-voucher
DELETE /buyer/checkout/sessions/{checkout_session_id}/voucher/{voucher_id}
POST   /buyer/checkout/sessions/{checkout_session_id}/validate
POST   /buyer/checkout/sessions/{checkout_session_id}/place-order
9. Order APIs - Buyer
GET    /buyer/orders
GET    /buyer/orders/{order_id}
GET    /buyer/orders/{order_id}/items
GET    /buyer/orders/{order_id}/status-history

POST   /buyer/orders/{order_id}/cancel
POST   /buyer/orders/{order_id}/confirm-received
POST   /buyer/orders/{order_id}/request-refund
POST   /buyer/orders/{order_id}/request-return
POST   /buyer/orders/{order_id}/dispute

GET    /buyer/orders/{order_id}/invoice
GET    /buyer/orders/{order_id}/tracking

Order filters:

?status=pending_payment|paid|processing|shipped|delivered|completed|cancelled|refunded|disputed
&page=
&limit=
10. Payment APIs
POST   /buyer/payments/create-intent
POST   /buyer/payments/confirm
GET    /buyer/payments/{payment_id}
POST   /buyer/payments/{payment_id}/retry
POST   /buyer/payments/{payment_id}/cancel

GET    /buyer/payment-methods
POST   /buyer/payment-methods
DELETE /buyer/payment-methods/{method_id}
POST   /buyer/payment-methods/{method_id}/set-default

Payment providers:

POST   /buyer/payments/stripe/create-intent
POST   /buyer/payments/paypal/create-order
POST   /buyer/payments/paypal/capture-order
POST   /buyer/payments/ecpay/create-order
POST   /buyer/payments/linepay/create-payment
POST   /buyer/payments/jkopay/create-payment
11. Shipping APIs - Buyer
GET    /buyer/shipping/options
GET    /buyer/shipments
GET    /buyer/shipments/{shipment_id}
GET    /buyer/shipments/{shipment_id}/tracking
12. Review APIs
POST   /buyer/reviews
GET    /buyer/reviews
GET    /buyer/reviews/{review_id}
PATCH  /buyer/reviews/{review_id}
DELETE /buyer/reviews/{review_id}

POST   /buyer/reviews/{review_id}/images
DELETE /buyer/reviews/{review_id}/images/{image_id}
POST   /buyer/reviews/{review_id}/vote
POST   /buyer/reviews/{review_id}/report
13. Voucher / Promotion APIs - Buyer
GET    /buyer/vouchers
GET    /buyer/vouchers/available
POST   /buyer/vouchers/claim
POST   /buyer/vouchers/validate
GET    /buyer/vouchers/{voucher_id}

GET    /public/promotions
GET    /public/promotions/{promotion_id}
GET    /public/campaigns
GET    /public/campaigns/{campaign_id}
GET    /public/flash-sales
GET    /public/flash-sales/{flash_sale_id}
14. Notification APIs
GET    /buyer/notifications
GET    /buyer/notifications/unread-count
PATCH  /buyer/notifications/{notification_id}/read
PATCH  /buyer/notifications/read-all
DELETE /buyer/notifications/{notification_id}

GET    /buyer/notification-preferences
PATCH  /buyer/notification-preferences

POST   /buyer/push-tokens
DELETE /buyer/push-tokens/{token_id}
15. Chat APIs - Buyer / Seller
GET    /chat/conversations
POST   /chat/conversations
GET    /chat/conversations/{conversation_id}
DELETE /chat/conversations/{conversation_id}

GET    /chat/conversations/{conversation_id}/messages
POST   /chat/conversations/{conversation_id}/messages
PATCH  /chat/conversations/{conversation_id}/read
POST   /chat/messages/{message_id}/attachments
DELETE /chat/messages/{message_id}

WebSocket:

/ws/chat
/ws/notifications
/ws/order-status
16. Seller Onboarding APIs
POST   /seller/apply
GET    /seller/application
PATCH  /seller/application
POST   /seller/documents
DELETE /seller/documents/{document_id}

GET    /seller/me
PATCH  /seller/me
PATCH  /seller/me/profile
PATCH  /seller/me/settings
POST   /seller/me/logo
POST   /seller/me/banner
17. Seller Dashboard APIs
GET    /seller/dashboard
GET    /seller/dashboard/stats
GET    /seller/dashboard/sales-chart
GET    /seller/dashboard/top-products
GET    /seller/dashboard/recent-orders
GET    /seller/dashboard/low-stock
18. Seller Product APIs
GET    /seller/products
POST   /seller/products
GET    /seller/products/{product_id}
PATCH  /seller/products/{product_id}
DELETE /seller/products/{product_id}

POST   /seller/products/{product_id}/publish
POST   /seller/products/{product_id}/unpublish
POST   /seller/products/{product_id}/duplicate

POST   /seller/products/{product_id}/images
PATCH  /seller/products/{product_id}/images/reorder
DELETE /seller/products/{product_id}/images/{image_id}

GET    /seller/products/{product_id}/variants
POST   /seller/products/{product_id}/variants
PATCH  /seller/products/{product_id}/variants/{variant_id}
DELETE /seller/products/{product_id}/variants/{variant_id}

POST   /seller/products/{product_id}/options
PATCH  /seller/products/{product_id}/options/{option_id}
DELETE /seller/products/{product_id}/options/{option_id}
19. Seller Inventory APIs
GET    /seller/inventory
GET    /seller/inventory/{inventory_item_id}
PATCH  /seller/inventory/{inventory_item_id}
POST   /seller/inventory/{inventory_item_id}/adjust
GET    /seller/inventory/{inventory_item_id}/movements
GET    /seller/inventory/low-stock
POST   /seller/inventory/bulk-update
20. Seller Order APIs
GET    /seller/orders
GET    /seller/orders/{order_id}
GET    /seller/orders/{order_id}/items

POST   /seller/orders/{order_id}/accept
POST   /seller/orders/{order_id}/reject
POST   /seller/orders/{order_id}/mark-processing
POST   /seller/orders/{order_id}/pack
POST   /seller/orders/{order_id}/ship
POST   /seller/orders/{order_id}/cancel

POST   /seller/orders/{order_id}/refund
POST   /seller/orders/{order_id}/return/approve
POST   /seller/orders/{order_id}/return/reject
21. Seller Shipping APIs
GET    /seller/shipping/settings
PATCH  /seller/shipping/settings

GET    /seller/shipments
GET    /seller/shipments/{shipment_id}
POST   /seller/shipments
PATCH  /seller/shipments/{shipment_id}
POST   /seller/shipments/{shipment_id}/print-label
POST   /seller/shipments/{shipment_id}/mark-shipped
GET    /seller/shipments/{shipment_id}/tracking
22. Seller Voucher / Promotion APIs
GET    /seller/vouchers
POST   /seller/vouchers
GET    /seller/vouchers/{voucher_id}
PATCH  /seller/vouchers/{voucher_id}
DELETE /seller/vouchers/{voucher_id}
POST   /seller/vouchers/{voucher_id}/activate
POST   /seller/vouchers/{voucher_id}/deactivate

GET    /seller/promotions
POST   /seller/promotions
PATCH  /seller/promotions/{promotion_id}
DELETE /seller/promotions/{promotion_id}
23. Seller Reviews APIs
GET    /seller/reviews
GET    /seller/reviews/{review_id}
POST   /seller/reviews/{review_id}/reply
PATCH  /seller/reviews/{review_id}/reply
DELETE /seller/reviews/{review_id}/reply
POST   /seller/reviews/{review_id}/report
24. Seller Wallet / Payout APIs
GET    /seller/wallet
GET    /seller/wallet/transactions
GET    /seller/payouts
POST   /seller/payouts/request
GET    /seller/payouts/{payout_id}

GET    /seller/bank-accounts
POST   /seller/bank-accounts
PATCH  /seller/bank-accounts/{bank_account_id}
DELETE /seller/bank-accounts/{bank_account_id}
POST   /seller/bank-accounts/{bank_account_id}/set-default
25. Seller Analytics APIs
GET    /seller/analytics/sales
GET    /seller/analytics/orders
GET    /seller/analytics/products
GET    /seller/analytics/customers
GET    /seller/analytics/traffic
GET    /seller/analytics/conversion
GET    /seller/analytics/refunds
26. Admin Auth / Dashboard APIs
POST   /admin/auth/login
POST   /admin/auth/logout
POST   /admin/auth/refresh-token

GET    /admin/dashboard
GET    /admin/dashboard/stats
GET    /admin/dashboard/sales-chart
GET    /admin/dashboard/order-chart
GET    /admin/dashboard/user-growth
GET    /admin/dashboard/seller-growth
27. Admin User Management APIs
GET    /admin/users
POST   /admin/users
GET    /admin/users/{user_id}
PATCH  /admin/users/{user_id}
DELETE /admin/users/{user_id}

POST   /admin/users/{user_id}/ban
POST   /admin/users/{user_id}/unban
POST   /admin/users/{user_id}/verify-email
POST   /admin/users/{user_id}/verify-phone
GET    /admin/users/{user_id}/orders
GET    /admin/users/{user_id}/activity
28. Admin Seller Management APIs
GET    /admin/sellers
GET    /admin/sellers/{seller_id}
PATCH  /admin/sellers/{seller_id}
POST   /admin/sellers/{seller_id}/approve
POST   /admin/sellers/{seller_id}/reject
POST   /admin/sellers/{seller_id}/suspend
POST   /admin/sellers/{seller_id}/unsuspend
GET    /admin/sellers/{seller_id}/products
GET    /admin/sellers/{seller_id}/orders
GET    /admin/sellers/{seller_id}/payouts
29. Admin Product Management APIs
GET    /admin/products
GET    /admin/products/{product_id}
PATCH  /admin/products/{product_id}
POST   /admin/products/{product_id}/approve
POST   /admin/products/{product_id}/reject
POST   /admin/products/{product_id}/ban
POST   /admin/products/{product_id}/unban
DELETE /admin/products/{product_id}

GET    /admin/reported-products
GET    /admin/reported-products/{report_id}
POST   /admin/reported-products/{report_id}/resolve
30. Admin Category / Brand APIs
GET    /admin/categories
POST   /admin/categories
GET    /admin/categories/{category_id}
PATCH  /admin/categories/{category_id}
DELETE /admin/categories/{category_id}
PATCH  /admin/categories/reorder

GET    /admin/brands
POST   /admin/brands
GET    /admin/brands/{brand_id}
PATCH  /admin/brands/{brand_id}
DELETE /admin/brands/{brand_id}
31. Admin Order APIs
GET    /admin/orders
GET    /admin/orders/{order_id}
PATCH  /admin/orders/{order_id}
POST   /admin/orders/{order_id}/cancel
POST   /admin/orders/{order_id}/refund
GET    /admin/orders/{order_id}/status-history
GET    /admin/orders/{order_id}/payments
GET    /admin/orders/{order_id}/shipments
32. Admin Payment APIs
GET    /admin/payments
GET    /admin/payments/{payment_id}
POST   /admin/payments/{payment_id}/refund
GET    /admin/payment-webhooks
GET    /admin/payment-webhooks/{webhook_event_id}
POST   /admin/payment-webhooks/{webhook_event_id}/reprocess
33. Admin Shipping APIs
GET    /admin/shipping-providers
POST   /admin/shipping-providers
PATCH  /admin/shipping-providers/{provider_id}
DELETE /admin/shipping-providers/{provider_id}

GET    /admin/shipments
GET    /admin/shipments/{shipment_id}
PATCH  /admin/shipments/{shipment_id}
GET    /admin/shipments/{shipment_id}/tracking
34. Admin Voucher / Campaign APIs
GET    /admin/vouchers
POST   /admin/vouchers
GET    /admin/vouchers/{voucher_id}
PATCH  /admin/vouchers/{voucher_id}
DELETE /admin/vouchers/{voucher_id}
POST   /admin/vouchers/{voucher_id}/activate
POST   /admin/vouchers/{voucher_id}/deactivate

GET    /admin/campaigns
POST   /admin/campaigns
GET    /admin/campaigns/{campaign_id}
PATCH  /admin/campaigns/{campaign_id}
DELETE /admin/campaigns/{campaign_id}

GET    /admin/flash-sales
POST   /admin/flash-sales
GET    /admin/flash-sales/{flash_sale_id}
PATCH  /admin/flash-sales/{flash_sale_id}
DELETE /admin/flash-sales/{flash_sale_id}
POST   /admin/flash-sales/{flash_sale_id}/items
DELETE /admin/flash-sales/{flash_sale_id}/items/{item_id}
35. Admin Reviews / Moderation APIs
GET    /admin/reviews
GET    /admin/reviews/{review_id}
DELETE /admin/reviews/{review_id}
POST   /admin/reviews/{review_id}/hide
POST   /admin/reviews/{review_id}/unhide

GET    /admin/reported-reviews
GET    /admin/reported-users
GET    /admin/reported-sellers
GET    /admin/moderation-cases
GET    /admin/moderation-cases/{case_id}
POST   /admin/moderation-cases/{case_id}/resolve
36. Admin CMS APIs
GET    /admin/cms/pages
POST   /admin/cms/pages
GET    /admin/cms/pages/{page_id}
PATCH  /admin/cms/pages/{page_id}
DELETE /admin/cms/pages/{page_id}

GET    /admin/cms/banners
POST   /admin/cms/banners
PATCH  /admin/cms/banners/{banner_id}
DELETE /admin/cms/banners/{banner_id}
PATCH  /admin/cms/banners/reorder

GET    /admin/cms/faqs
POST   /admin/cms/faqs
PATCH  /admin/cms/faqs/{faq_id}
DELETE /admin/cms/faqs/{faq_id}

GET    /admin/cms/announcements
POST   /admin/cms/announcements
PATCH  /admin/cms/announcements/{announcement_id}
DELETE /admin/cms/announcements/{announcement_id}
37. Admin RBAC APIs
GET    /admin/roles
POST   /admin/roles
GET    /admin/roles/{role_id}
PATCH  /admin/roles/{role_id}
DELETE /admin/roles/{role_id}

GET    /admin/permissions
POST   /admin/permissions

GET    /admin/roles/{role_id}/permissions
POST   /admin/roles/{role_id}/permissions
DELETE /admin/roles/{role_id}/permissions/{permission_id}

POST   /admin/users/{user_id}/roles
DELETE /admin/users/{user_id}/roles/{role_id}
38. Admin Settings APIs
GET    /admin/settings
PATCH  /admin/settings

GET    /admin/feature-flags
POST   /admin/feature-flags
PATCH  /admin/feature-flags/{flag_id}
DELETE /admin/feature-flags/{flag_id}

GET    /admin/audit-logs
GET    /admin/activity-logs
GET    /admin/system-health
39. Media Upload APIs
POST   /media/upload
POST   /media/upload/presigned-url
GET    /media/files
GET    /media/files/{file_id}
DELETE /media/files/{file_id}

POST   /seller/media/upload
POST   /admin/media/upload
40. Webhook APIs
POST   /webhooks/stripe
POST   /webhooks/paypal
POST   /webhooks/ecpay
POST   /webhooks/linepay
POST   /webhooks/jkopay

POST   /webhooks/shipping/{provider}
POST   /webhooks/sms/{provider}
POST   /webhooks/email/{provider}
41. Internal Worker APIs

These are private/internal only.

POST   /internal/jobs/send-email
POST   /internal/jobs/send-sms
POST   /internal/jobs/send-push
POST   /internal/jobs/index-product
POST   /internal/jobs/reindex-products
POST   /internal/jobs/process-image
POST   /internal/jobs/process-video
POST   /internal/jobs/expire-checkout
POST   /internal/jobs/cancel-unpaid-orders
POST   /internal/jobs/release-expired-inventory
POST   /internal/jobs/update-shipment-tracking
POST   /internal/jobs/settle-seller-wallets
POST   /internal/jobs/process-payouts
POST   /internal/jobs/generate-reports
42. Health / System APIs
GET    /health
GET    /ready
GET    /metrics
GET    /version
MVP API list to build first

Start with these:

POST   /auth/register
POST   /auth/login
POST   /auth/refresh-token
GET    /users/me
PATCH  /users/me

GET    /public/categories
GET    /public/products
GET    /public/products/{product_id}
GET    /public/search

POST   /seller/apply
GET    /seller/me
PATCH  /seller/me
POST   /seller/products
GET    /seller/products
PATCH  /seller/products/{product_id}
POST   /seller/products/{product_id}/variants
PATCH  /seller/inventory/{inventory_item_id}

GET    /buyer/cart
POST   /buyer/cart/items
PATCH  /buyer/cart/items/{cart_item_id}
DELETE /buyer/cart/items/{cart_item_id}

POST   /buyer/checkout/sessions
POST   /buyer/checkout/sessions/{checkout_session_id}/place-order

GET    /buyer/orders
GET    /buyer/orders/{order_id}

POST   /buyer/payments/create-intent
POST   /webhooks/stripe

GET    /seller/orders
POST   /seller/orders/{order_id}/ship

POST   /buyer/reviews

GET    /buyer/notifications
GET    /admin/dashboard
GET    /admin/users
GET    /admin/sellers
GET    /admin/products
GET    /admin/orders

This gives you the full API surface for buyer app, seller center, admin panel, payments, shipping, chat, reviews, vouchers, moderation, and backend operations.