# Shopee-like Ecommerce Marketplace

## 1. Users & Auth
- `users`
- `user_profiles`
- `user_addresses`
- `user_sessions`
- `user_devices`
- `user_oauth_accounts`
- `user_phone_verifications`
- `user_email_verifications`
- `password_reset_tokens`
- `refresh_tokens`
- `login_attempts`

## 2. Roles & Permissions
- `roles`
- `permissions`
- `role_permissions`
- `user_roles`
- `admin_users`
- `admin_activity_logs`

## 3. Sellers / Shops
- `sellers`
- `seller_profiles`
- `seller_documents`
- `seller_bank_accounts`
- `seller_wallets`
- `seller_wallet_transactions`
- `seller_payouts`
- `seller_store_settings`
- `seller_followers`
- `seller_ratings`

## 4. Product Catalog
- `products`
- `product_translations`
- `product_images`
- `product_videos`
- `product_categories`
- `categories`
- `category_translations`
- `brands`
- `brand_translations`
- `product_variants`
- `variant_options`
- `variant_option_values`
- `product_attributes`
- `attribute_values`
- `product_specs`
- `product_tags`
- `tags`
- `product_collections`
- `collection_products`

## 5. Inventory
- `inventory_items`
- `inventory_movements`
- `inventory_reservations`
- `stock_locations`
- `warehouse_locations`
- `low_stock_alerts`

## 6. Cart & Wishlist
- `carts`
- `cart_items`
- `wishlists`
- `wishlist_items`
- `recently_viewed_products`
- `saved_for_later_items`

## 7. Checkout
- `checkout_sessions`
- `checkout_items`
- `checkout_addresses`
- `checkout_payment_methods`
- `checkout_shipping_options`
- `checkout_discounts`
- `checkout_price_snapshots`

## 8. Orders
- `orders`
- `order_items`
- `order_item_snapshots`
- `order_status_history`
- `order_notes`
- `order_cancellations`
- `order_returns`
- `order_return_items`
- `order_refunds`
- `order_refund_items`
- `order_disputes`
- `order_dispute_messages`

## 9. Payments
- `payments`
- `payment_attempts`
- `payment_methods`
- `payment_transactions`
- `payment_webhook_events`
- `payment_refunds`
- `payment_provider_accounts`
- `payment_idempotency_keys`

## 10. Shipping & Logistics
- `shipping_methods`
- `shipping_providers`
- `shipping_zones`
- `shipping_rates`
- `shipments`
- `shipment_items`
- `shipment_tracking_events`
- `delivery_attempts`
- `pickup_points`
- `seller_shipping_settings`

## 11. Promotions / Vouchers
- `promotions`
- `promotion_rules`
- `promotion_conditions`
- `promotion_rewards`
- `vouchers`
- `voucher_codes`
- `voucher_redemptions`
- `flash_sales`
- `flash_sale_items`
- `campaigns`
- `campaign_products`
- `free_shipping_rules`
- `coin_reward_rules`

## 12. Reviews & Ratings
- `product_reviews`
- `product_review_images`
- `product_review_votes`
- `seller_reviews`
- `review_replies`
- `review_reports`

## 13. Search & Discovery
- `search_queries`
- `search_clicks`
- `search_indexes`
- `product_recommendations`
- `user_recommendation_events`
- `trending_products`
- `homepage_sections`
- `homepage_section_items`

## 14. Chat / Messaging
- `chat_conversations`
- `chat_participants`
- `chat_messages`
- `chat_message_attachments`
- `chat_read_receipts`
- `chat_blocks`

## 15. Notifications
- `notifications`
- `notification_templates`
- `notification_preferences`
- `push_tokens`
- `email_logs`
- `sms_logs`
- `in_app_notification_reads`

## 16. Media / Files
- `media_files`
- `media_folders`
- `media_usage`
- `image_processing_jobs`
- `video_processing_jobs`

## 17. Wallet / Coins / Rewards
- `user_wallets`
- `wallet_transactions`
- `coins`
- `coin_transactions`
- `reward_points`
- `reward_point_transactions`
- `referrals`
- `referral_rewards`

## 18. Buyer Protection / Disputes
- `buyer_protection_cases`
- `case_messages`
- `case_evidence_files`
- `case_status_history`
- `case_resolutions`

## 19. Moderation
- `reported_products`
- `reported_reviews`
- `reported_users`
- `reported_sellers`
- `moderation_cases`
- `moderation_actions`
- `blocked_keywords`
- `content_moderation_logs`

## 20. Admin / CMS
- `cms_pages`
- `cms_banners`
- `cms_banner_translations`
- `faq_categories`
- `faqs`
- `faq_translations`
- `announcements`
- `announcement_translations`
- `system_settings`
- `feature_flags`

## 21. Analytics / Events
- `event_logs`
- `user_events`
- `product_events`
- `order_events`
- `seller_events`
- `payment_events`
- `traffic_sources`
- `conversion_funnels`
- `daily_sales_stats`
- `seller_daily_stats`
- `product_daily_stats`

## 22. Audit & System
- `audit_logs`
- `webhook_events`
- `background_jobs`
- `job_attempts`
- `api_keys`
- `rate_limits`
- `idempotency_keys`
- `system_health_checks`

## Most Important MVP Tables

Start with these first:

- `users`
- `user_addresses`
- `sellers`
- `products`
- `product_images`
- `categories`
- `product_variants`
- `inventory_items`
- `carts`
- `cart_items`
- `orders`
- `order_items`
- `payments`
- `shipments`
- `vouchers`
- `product_reviews`
- `notifications`
- `admin_users`
- `roles`
- `permissions`
- `audit_logs`