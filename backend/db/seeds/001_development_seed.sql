-- Repeatable development/staging seed data.
INSERT INTO roles (id, name, description)
VALUES
  ('00000000-0000-0000-0000-000000000101', 'admin', 'Full platform administrator'),
  ('00000000-0000-0000-0000-000000000102', 'seller', 'Marketplace seller'),
  ('00000000-0000-0000-0000-000000000103', 'buyer', 'Marketplace buyer'),
  ('00000000-0000-0000-0000-000000000104', 'ops_payments', 'Ops payments approver (Wave/manual payouts)')
ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description;

INSERT INTO permissions (id, key, description)
VALUES
  ('00000000-0000-0000-0000-000000000201', 'admin.dashboard.read', 'Read admin dashboard'),
  ('00000000-0000-0000-0000-000000000202', 'admin.users.manage', 'Manage users'),
  ('00000000-0000-0000-0000-000000000203', 'admin.sellers.manage', 'Manage sellers'),
  ('00000000-0000-0000-0000-000000000204', 'admin.products.manage', 'Manage products'),
  ('00000000-0000-0000-0000-000000000205', 'admin.orders.manage', 'Manage orders'),
  ('00000000-0000-0000-0000-000000000209', 'admin.payments.manage', 'Manage payments approvals/refunds'),
  ('00000000-0000-0000-0000-000000000206', 'seller.products.manage', 'Manage seller products'),
  ('00000000-0000-0000-0000-000000000207', 'seller.orders.manage', 'Manage seller orders'),
  ('00000000-0000-0000-0000-000000000208', 'buyer.orders.read', 'Read own orders')
ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('seller.products.manage', 'seller.orders.manage')
WHERE r.name = 'seller'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('buyer.orders.read')
WHERE r.name = 'buyer'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.key IN ('admin.payments.manage')
WHERE r.name = 'ops_payments'
ON CONFLICT DO NOTHING;

INSERT INTO users (id, email, phone, password_hash, status, email_verified, phone_verified)
VALUES
  ('00000000-0000-0000-0000-000000001001', 'admin@bitik.local', '+10000000001', '$2a$10$xbra1pjJj4jk3D1MLcDcuOyphbdWbzP8SNzLfScLQDbjse0X2yVfG', 'active', TRUE, TRUE),
  ('00000000-0000-0000-0000-000000001002', 'seller@bitik.local', '+10000000002', '$2a$10$xbra1pjJj4jk3D1MLcDcuOyphbdWbzP8SNzLfScLQDbjse0X2yVfG', 'active', TRUE, TRUE),
  ('00000000-0000-0000-0000-000000001003', 'buyer@bitik.local', '+10000000003', '$2a$10$xbra1pjJj4jk3D1MLcDcuOyphbdWbzP8SNzLfScLQDbjse0X2yVfG', 'active', TRUE, TRUE)
ON CONFLICT (email) DO UPDATE
SET status = EXCLUDED.status,
    email_verified = EXCLUDED.email_verified,
    phone_verified = EXCLUDED.phone_verified;

INSERT INTO user_profiles (user_id, display_name, first_name, last_name, language, country_code)
VALUES
  ('00000000-0000-0000-0000-000000001001', 'Bitik Admin', 'Bitik', 'Admin', 'en', 'US'),
  ('00000000-0000-0000-0000-000000001002', 'Demo Seller', 'Demo', 'Seller', 'en', 'US'),
  ('00000000-0000-0000-0000-000000001003', 'Demo Buyer', 'Demo', 'Buyer', 'en', 'US')
ON CONFLICT (user_id) DO UPDATE
SET display_name = EXCLUDED.display_name,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    language = EXCLUDED.language,
    country_code = EXCLUDED.country_code;

INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000001001', id FROM roles WHERE name = 'admin'
ON CONFLICT DO NOTHING;
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000001002', id FROM roles WHERE name = 'seller'
ON CONFLICT DO NOTHING;
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000001003', id FROM roles WHERE name = 'buyer'
ON CONFLICT DO NOTHING;

INSERT INTO categories (id, name, slug, sort_order, is_active)
VALUES
  ('00000000-0000-0000-0000-000000002001', 'Fashion', 'fashion', 10, TRUE),
  ('00000000-0000-0000-0000-000000002002', 'Electronics', 'electronics', 20, TRUE),
  ('00000000-0000-0000-0000-000000002003', 'Home & Living', 'home-living', 30, TRUE)
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    sort_order = EXCLUDED.sort_order,
    is_active = EXCLUDED.is_active;

INSERT INTO brands (id, name, slug, is_active)
VALUES
  ('00000000-0000-0000-0000-000000003001', 'Bitik Basics', 'bitik-basics', TRUE),
  ('00000000-0000-0000-0000-000000003002', 'Demo Gear', 'demo-gear', TRUE)
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    is_active = EXCLUDED.is_active;

INSERT INTO seller_applications (id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at)
VALUES (
  '00000000-0000-0000-0000-000000004001',
  '00000000-0000-0000-0000-000000001002',
  'Demo Seller Shop',
  'demo-seller-shop',
  'individual',
  'US',
  'USD',
  'approved',
  now(),
  '00000000-0000-0000-0000-000000001001',
  now()
)
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    reviewed_by = EXCLUDED.reviewed_by,
    reviewed_at = EXCLUDED.reviewed_at;

INSERT INTO sellers (id, user_id, application_id, shop_name, slug, description, status)
VALUES (
  '00000000-0000-0000-0000-000000005001',
  '00000000-0000-0000-0000-000000001002',
  '00000000-0000-0000-0000-000000004001',
  'Demo Seller Shop',
  'demo-seller-shop',
  'Seed seller for local development.',
  'active'
)
ON CONFLICT (slug) DO UPDATE
SET shop_name = EXCLUDED.shop_name,
    description = EXCLUDED.description,
    status = EXCLUDED.status;

INSERT INTO seller_wallets (seller_id, currency)
VALUES ('00000000-0000-0000-0000-000000005001', 'USD')
ON CONFLICT (seller_id) DO NOTHING;

INSERT INTO shipping_providers (id, name, code, metadata, is_active)
VALUES
  ('00000000-0000-0000-0000-000000006001', 'Local Courier', 'local-courier', '{"tracking_url_template":"https://tracking.local/{tracking_number}"}', TRUE),
  ('00000000-0000-0000-0000-000000006002', 'Pickup', 'pickup', '{}', TRUE)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    metadata = EXCLUDED.metadata,
    is_active = EXCLUDED.is_active;

INSERT INTO platform_settings (key, value, description, is_public)
VALUES
  ('default_currency', '"USD"', 'Default marketplace currency.', TRUE),
  ('supported_languages', '["en"]', 'Supported language codes.', TRUE),
  ('manual_wave_enabled', 'true', 'Enable manual Wave payment approvals.', FALSE),
  ('wave_manual_instructions', '"Send the exact amount via Wave and include your order number in the reference. After you submit, our team will verify and approve your payment."', 'Buyer-facing Wave manual instructions.', TRUE),
  ('pod_enabled', 'true', 'Enable pay on delivery.', FALSE)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    description = EXCLUDED.description,
    is_public = EXCLUDED.is_public;

INSERT INTO feature_flags (key, description, enabled, rules)
VALUES
  ('checkout_v1', 'Enable checkout v1 APIs.', TRUE, '{}'),
  ('seller_onboarding_v1', 'Enable seller onboarding v1 APIs.', TRUE, '{}')
ON CONFLICT (key) DO UPDATE
SET description = EXCLUDED.description,
    enabled = EXCLUDED.enabled,
    rules = EXCLUDED.rules;

INSERT INTO products (id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, published_at)
VALUES
  (
    '00000000-0000-0000-0000-000000007001',
    '00000000-0000-0000-0000-000000005001',
    '00000000-0000-0000-0000-000000002001',
    '00000000-0000-0000-0000-000000003001',
    'Bitik Basic T-Shirt',
    'bitik-basic-t-shirt',
    'Seed product for local catalog development.',
    'active',
    1999,
    1999,
    'USD',
    now()
  ),
  (
    '00000000-0000-0000-0000-000000007002',
    '00000000-0000-0000-0000-000000005001',
    '00000000-0000-0000-0000-000000002002',
    '00000000-0000-0000-0000-000000003002',
    'Demo Wireless Earbuds',
    'demo-wireless-earbuds',
    'Seed electronics product for local catalog development.',
    'active',
    4999,
    4999,
    'USD',
    now()
  )
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    min_price_cents = EXCLUDED.min_price_cents,
    max_price_cents = EXCLUDED.max_price_cents,
    published_at = EXCLUDED.published_at;

INSERT INTO product_variants (id, product_id, sku, name, price_cents, currency, is_active)
VALUES
  ('00000000-0000-0000-0000-000000008001', '00000000-0000-0000-0000-000000007001', 'BITIK-TSHIRT-M', 'Medium', 1999, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008002', '00000000-0000-0000-0000-000000007002', 'DEMO-EARBUDS-WHT', 'White', 4999, 'USD', TRUE)
ON CONFLICT (sku) DO UPDATE
SET name = EXCLUDED.name,
    price_cents = EXCLUDED.price_cents,
    is_active = EXCLUDED.is_active;

INSERT INTO inventory_items (product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold)
VALUES
  ('00000000-0000-0000-0000-000000007001', '00000000-0000-0000-0000-000000008001', 100, 0, 10),
  ('00000000-0000-0000-0000-000000007002', '00000000-0000-0000-0000-000000008002', 50, 0, 5)
ON CONFLICT (variant_id) DO UPDATE
SET quantity_available = EXCLUDED.quantity_available,
    low_stock_threshold = EXCLUDED.low_stock_threshold;

INSERT INTO product_images (id, product_id, url, alt_text, sort_order, is_primary)
VALUES
  ('00000000-0000-0000-0000-000000009001', '00000000-0000-0000-0000-000000007001', 'https://placehold.co/900x1200?text=Bitik+T-Shirt', 'Bitik Basic T-Shirt', 10, TRUE),
  ('00000000-0000-0000-0000-000000009002', '00000000-0000-0000-0000-000000007002', 'https://placehold.co/900x900?text=Demo+Earbuds', 'Demo Wireless Earbuds', 10, TRUE)
ON CONFLICT (id) DO UPDATE
SET url = EXCLUDED.url,
    alt_text = EXCLUDED.alt_text,
    sort_order = EXCLUDED.sort_order,
    is_primary = EXCLUDED.is_primary;

INSERT INTO product_reviews (id, product_id, user_id, rating, title, body, is_verified_purchase)
VALUES
  ('00000000-0000-0000-0000-000000010001', '00000000-0000-0000-0000-000000007001', '00000000-0000-0000-0000-000000001003', 5, 'Great basic', 'Comfortable seed catalog item.', TRUE),
  ('00000000-0000-0000-0000-000000010002', '00000000-0000-0000-0000-000000007002', '00000000-0000-0000-0000-000000001003', 4, 'Solid demo gear', 'Useful for local catalog testing.', TRUE)
ON CONFLICT (id) DO UPDATE
SET rating = EXCLUDED.rating,
    title = EXCLUDED.title,
    body = EXCLUDED.body,
    is_verified_purchase = EXCLUDED.is_verified_purchase;

INSERT INTO cms_banners (id, title, image_url, link_url, placement, sort_order, status, starts_at)
VALUES
  ('00000000-0000-0000-0000-000000011001', 'New Season Picks', 'https://placehold.co/1600x600?text=Bitik+New+Season', '/public/products?sort=newest', 'home', 10, 'published', now()),
  ('00000000-0000-0000-0000-000000011002', 'Top Electronics', 'https://placehold.co/1600x600?text=Top+Electronics', '/public/categories/00000000-0000-0000-0000-000000002002/products', 'home', 20, 'published', now())
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    image_url = EXCLUDED.image_url,
    link_url = EXCLUDED.link_url,
    placement = EXCLUDED.placement,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    starts_at = EXCLUDED.starts_at;

INSERT INTO platform_settings (key, value, description, is_public)
VALUES
  ('home.featured_categories', '["fashion","electronics","home-living"]', 'Homepage featured category slugs.', TRUE),
  ('home.featured_products', '["bitik-basic-t-shirt","demo-wireless-earbuds"]', 'Homepage featured product slugs.', TRUE),
  ('home.layout', '{"sections":["banners","featured_categories","featured_products"]}', 'Homepage section order.', TRUE)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    description = EXCLUDED.description,
    is_public = EXCLUDED.is_public;
