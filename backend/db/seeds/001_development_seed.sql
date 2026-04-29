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
  ('00000000-0000-0000-0000-000000001003', 'buyer@bitik.local', '+10000000003', '$2a$10$xbra1pjJj4jk3D1MLcDcuOyphbdWbzP8SNzLfScLQDbjse0X2yVfG', 'active', TRUE, TRUE),
  ('00000000-0000-0000-0000-000000001004', 'seller.home@bitik.local', '+10000000004', '$2a$10$xbra1pjJj4jk3D1MLcDcuOyphbdWbzP8SNzLfScLQDbjse0X2yVfG', 'active', TRUE, TRUE),
  ('00000000-0000-0000-0000-000000001005', 'seller.beauty@bitik.local', '+10000000005', '$2a$10$xbra1pjJj4jk3D1MLcDcuOyphbdWbzP8SNzLfScLQDbjse0X2yVfG', 'active', TRUE, TRUE)
ON CONFLICT (email) DO UPDATE
SET status = EXCLUDED.status,
    email_verified = EXCLUDED.email_verified,
    phone_verified = EXCLUDED.phone_verified;

INSERT INTO user_profiles (user_id, display_name, first_name, last_name, language, country_code)
VALUES
  ('00000000-0000-0000-0000-000000001001', 'Bitik Admin', 'Bitik', 'Admin', 'en', 'US'),
  ('00000000-0000-0000-0000-000000001002', 'Nora Atelier', 'Nora', 'Atelier', 'en', 'US'),
  ('00000000-0000-0000-0000-000000001003', 'Demo Buyer', 'Demo', 'Buyer', 'en', 'US'),
  ('00000000-0000-0000-0000-000000001004', 'Casa Market', 'Casa', 'Market', 'en', 'US'),
  ('00000000-0000-0000-0000-000000001005', 'Glow Lab', 'Glow', 'Lab', 'en', 'US')
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
SELECT '00000000-0000-0000-0000-000000001004', id FROM roles WHERE name = 'seller'
ON CONFLICT DO NOTHING;
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000001005', id FROM roles WHERE name = 'seller'
ON CONFLICT DO NOTHING;
INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-000000001003', id FROM roles WHERE name = 'buyer'
ON CONFLICT DO NOTHING;

INSERT INTO categories (id, name, slug, image_url, sort_order, is_active)
VALUES
  ('00000000-0000-0000-0000-000000002001', 'Fashion', 'fashion', 'https://images.unsplash.com/photo-1483985988355-763728e1935b?auto=format&fit=crop&w=1200&q=80', 10, TRUE),
  ('00000000-0000-0000-0000-000000002002', 'Electronics', 'electronics', 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?auto=format&fit=crop&w=1200&q=80', 20, TRUE),
  ('00000000-0000-0000-0000-000000002003', 'Home & Living', 'home-living', 'https://images.unsplash.com/photo-1513161455079-7dc1de15ef3e?auto=format&fit=crop&w=1200&q=80', 30, TRUE),
  ('00000000-0000-0000-0000-000000002004', 'Beauty', 'beauty', 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?auto=format&fit=crop&w=1200&q=80', 40, TRUE),
  ('00000000-0000-0000-0000-000000002005', 'Bags & Travel', 'bags-travel', 'https://images.unsplash.com/photo-1553062407-98eeb64c6a62?auto=format&fit=crop&w=1200&q=80', 50, TRUE),
  ('00000000-0000-0000-0000-000000002006', 'Kitchen', 'kitchen', 'https://images.unsplash.com/photo-1556911220-bff31c812dba?auto=format&fit=crop&w=1200&q=80', 60, TRUE)
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    image_url = EXCLUDED.image_url,
    sort_order = EXCLUDED.sort_order,
    is_active = EXCLUDED.is_active;

INSERT INTO brands (id, name, slug, logo_url, is_active)
VALUES
  ('00000000-0000-0000-0000-000000003001', 'Nora Atelier', 'nora-atelier', 'https://images.unsplash.com/photo-1529139574466-a303027c1d8b?auto=format&fit=crop&w=600&q=80', TRUE),
  ('00000000-0000-0000-0000-000000003002', 'Pulse Audio', 'pulse-audio', 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=600&q=80', TRUE),
  ('00000000-0000-0000-0000-000000003003', 'Casa Studio', 'casa-studio', 'https://images.unsplash.com/photo-1493663284031-b7e3aefcae8e?auto=format&fit=crop&w=600&q=80', TRUE),
  ('00000000-0000-0000-0000-000000003004', 'Glow Lab', 'glow-lab', 'https://images.unsplash.com/photo-1570172619644-dfd03ed5d881?auto=format&fit=crop&w=600&q=80', TRUE),
  ('00000000-0000-0000-0000-000000003005', 'Carry Co.', 'carry-co', 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=600&q=80', TRUE),
  ('00000000-0000-0000-0000-000000003006', 'Mellow Kitchen', 'mellow-kitchen', 'https://images.unsplash.com/photo-1556909114-f6e7ad7d3136?auto=format&fit=crop&w=600&q=80', TRUE)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    logo_url = EXCLUDED.logo_url,
    is_active = EXCLUDED.is_active;

INSERT INTO seller_applications (id, user_id, shop_name, slug, business_type, country, currency, status, submitted_at, reviewed_by, reviewed_at)
VALUES
  ('00000000-0000-0000-0000-000000004001', '00000000-0000-0000-0000-000000001002', 'Nora Atelier', 'nora-atelier-shop', 'individual', 'US', 'USD', 'approved', now(), '00000000-0000-0000-0000-000000001001', now()),
  ('00000000-0000-0000-0000-000000004002', '00000000-0000-0000-0000-000000001004', 'Casa Market', 'casa-market', 'company', 'US', 'USD', 'approved', now(), '00000000-0000-0000-0000-000000001001', now()),
  ('00000000-0000-0000-0000-000000004003', '00000000-0000-0000-0000-000000001005', 'Glow Lab', 'glow-lab-shop', 'individual', 'US', 'USD', 'approved', now(), '00000000-0000-0000-0000-000000001001', now())
ON CONFLICT (id) DO UPDATE
SET user_id = EXCLUDED.user_id,
    shop_name = EXCLUDED.shop_name,
    slug = EXCLUDED.slug,
    business_type = EXCLUDED.business_type,
    country = EXCLUDED.country,
    currency = EXCLUDED.currency,
    status = EXCLUDED.status,
    reviewed_by = EXCLUDED.reviewed_by,
    reviewed_at = EXCLUDED.reviewed_at;

INSERT INTO sellers (id, user_id, application_id, shop_name, slug, description, status)
VALUES
  ('00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000001002', '00000000-0000-0000-0000-000000004001', 'Nora Atelier', 'nora-atelier-shop', 'Minimal wardrobe staples, made for everyday wear.', 'active'),
  ('00000000-0000-0000-0000-000000005002', '00000000-0000-0000-0000-000000001004', '00000000-0000-0000-0000-000000004002', 'Casa Market', 'casa-market', 'Warm home, kitchen, and travel essentials.', 'active'),
  ('00000000-0000-0000-0000-000000005003', '00000000-0000-0000-0000-000000001005', '00000000-0000-0000-0000-000000004003', 'Glow Lab', 'glow-lab-shop', 'Clean beauty and daily care picks.', 'active')
ON CONFLICT (id) DO UPDATE
SET user_id = EXCLUDED.user_id,
    application_id = EXCLUDED.application_id,
    shop_name = EXCLUDED.shop_name,
    slug = EXCLUDED.slug,
    description = EXCLUDED.description,
    status = EXCLUDED.status;

INSERT INTO seller_wallets (seller_id, currency)
VALUES
  ('00000000-0000-0000-0000-000000005001', 'USD'),
  ('00000000-0000-0000-0000-000000005002', 'USD'),
  ('00000000-0000-0000-0000-000000005003', 'USD')
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

INSERT INTO products (id, seller_id, category_id, brand_id, name, slug, description, status, min_price_cents, max_price_cents, currency, total_sold, rating, review_count, published_at)
VALUES
  ('00000000-0000-0000-0000-000000007001', '00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000002001', '00000000-0000-0000-0000-000000003001', 'Organic Cotton Relaxed Tee', 'organic-cotton-relaxed-tee', 'Midweight organic cotton tee with a soft drape and clean neckline.', 'active', 2400, 2800, 'USD', 842, 4.80, 128, now()),
  ('00000000-0000-0000-0000-000000007002', '00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000002001', '00000000-0000-0000-0000-000000003001', 'Linen Overshirt', 'linen-overshirt-sand', 'Breathable linen overshirt with corozo buttons and a relaxed fit.', 'active', 6400, 7200, 'USD', 391, 4.70, 84, now()),
  ('00000000-0000-0000-0000-000000007003', '00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000002005', '00000000-0000-0000-0000-000000003005', 'Everyday Canvas Tote', 'everyday-canvas-tote', 'Structured canvas tote with padded handles and an inside zip pocket.', 'active', 3800, 3800, 'USD', 612, 4.90, 76, now()),
  ('00000000-0000-0000-0000-000000007004', '00000000-0000-0000-0000-000000005002', '00000000-0000-0000-0000-000000002003', '00000000-0000-0000-0000-000000003003', 'Stoneware Dinner Set', 'stoneware-dinner-set-cream', 'Four-piece glazed stoneware set with plates, bowls, and mugs.', 'active', 8900, 8900, 'USD', 238, 4.60, 55, now()),
  ('00000000-0000-0000-0000-000000007005', '00000000-0000-0000-0000-000000005002', '00000000-0000-0000-0000-000000002006', '00000000-0000-0000-0000-000000003006', 'Matte Pour-Over Kettle', 'matte-pour-over-kettle', 'Stainless kettle with gooseneck spout and heat-safe handle.', 'active', 5600, 5600, 'USD', 327, 4.80, 91, now()),
  ('00000000-0000-0000-0000-000000007006', '00000000-0000-0000-0000-000000005002', '00000000-0000-0000-0000-000000002003', '00000000-0000-0000-0000-000000003003', 'Textured Cotton Throw', 'textured-cotton-throw', 'Soft woven cotton throw for sofas, beds, and cool evenings.', 'active', 4800, 5200, 'USD', 486, 4.70, 69, now()),
  ('00000000-0000-0000-0000-000000007007', '00000000-0000-0000-0000-000000005003', '00000000-0000-0000-0000-000000002004', '00000000-0000-0000-0000-000000003004', 'Hydrating Serum Duo', 'hydrating-serum-duo', 'Lightweight hyaluronic serum and barrier oil for daily skincare.', 'active', 4200, 4200, 'USD', 954, 4.90, 142, now()),
  ('00000000-0000-0000-0000-000000007008', '00000000-0000-0000-0000-000000005003', '00000000-0000-0000-0000-000000002004', '00000000-0000-0000-0000-000000003004', 'Mineral Sunscreen SPF 50', 'mineral-sunscreen-spf-50', 'Non-greasy mineral sunscreen with no white cast for daily use.', 'active', 2600, 2600, 'USD', 1172, 4.70, 204, now()),
  ('00000000-0000-0000-0000-000000007009', '00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000002001', '00000000-0000-0000-0000-000000003001', 'Slim Rib Tank', 'slim-rib-tank-black', 'Stretch rib tank with clean binding and a close fit.', 'active', 2200, 2200, 'USD', 705, 4.60, 98, now()),
  ('00000000-0000-0000-0000-000000007010', '00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000002005', '00000000-0000-0000-0000-000000003005', 'Compact Travel Backpack', 'compact-travel-backpack', 'Water-resistant backpack with laptop sleeve and quick-access pockets.', 'active', 7800, 8400, 'USD', 519, 4.80, 121, now()),
  ('00000000-0000-0000-0000-000000007011', '00000000-0000-0000-0000-000000005002', '00000000-0000-0000-0000-000000002002', '00000000-0000-0000-0000-000000003002', 'Noise Cancelling Headphones', 'noise-cancelling-headphones', 'Over-ear wireless headphones with active noise cancellation and 40-hour battery.', 'active', 12900, 14900, 'USD', 284, 4.80, 63, now()),
  ('00000000-0000-0000-0000-000000007012', '00000000-0000-0000-0000-000000005002', '00000000-0000-0000-0000-000000002002', '00000000-0000-0000-0000-000000003002', 'Wireless Charging Stand', 'wireless-charging-stand', 'Aluminum magnetic charging stand for phone and earbuds.', 'active', 3600, 3600, 'USD', 744, 4.50, 88, now()),
  ('00000000-0000-0000-0000-000000007013', '00000000-0000-0000-0000-000000005003', '00000000-0000-0000-0000-000000002004', '00000000-0000-0000-0000-000000003004', 'Clean Lip Tint Set', 'clean-lip-tint-set', 'Three sheer lip tints with a balmy finish.', 'active', 3200, 3200, 'USD', 631, 4.70, 93, now()),
  ('00000000-0000-0000-0000-000000007014', '00000000-0000-0000-0000-000000005002', '00000000-0000-0000-0000-000000002006', '00000000-0000-0000-0000-000000003006', 'Acacia Cutting Board', 'acacia-cutting-board', 'Reversible acacia board with juice groove and soft-rounded edges.', 'active', 3400, 3400, 'USD', 452, 4.80, 77, now())
ON CONFLICT (id) DO UPDATE
SET seller_id = EXCLUDED.seller_id,
    category_id = EXCLUDED.category_id,
    brand_id = EXCLUDED.brand_id,
    name = EXCLUDED.name,
    slug = EXCLUDED.slug,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    min_price_cents = EXCLUDED.min_price_cents,
    max_price_cents = EXCLUDED.max_price_cents,
    total_sold = EXCLUDED.total_sold,
    rating = EXCLUDED.rating,
    review_count = EXCLUDED.review_count,
    published_at = EXCLUDED.published_at;

INSERT INTO product_variants (id, product_id, sku, name, price_cents, currency, is_active)
VALUES
  ('00000000-0000-0000-0000-000000008001', '00000000-0000-0000-0000-000000007001', 'TEE-ORG-S', 'Small', 2400, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008002', '00000000-0000-0000-0000-000000007001', 'TEE-ORG-M', 'Medium', 2600, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008003', '00000000-0000-0000-0000-000000007002', 'LINEN-OVR-SAND-M', 'Sand / M', 6800, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008004', '00000000-0000-0000-0000-000000007003', 'TOTE-NATURAL', 'Natural', 3800, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008005', '00000000-0000-0000-0000-000000007004', 'STONEWARE-CREAM-4PC', 'Cream / 4 pc', 8900, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008006', '00000000-0000-0000-0000-000000007005', 'KETTLE-MATTE-BLACK', 'Matte black', 5600, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008007', '00000000-0000-0000-0000-000000007006', 'THROW-OAT', 'Oat', 4800, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008008', '00000000-0000-0000-0000-000000007007', 'SERUM-DUO-30ML', '30 ml duo', 4200, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008009', '00000000-0000-0000-0000-000000007008', 'SPF50-50ML', '50 ml', 2600, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008010', '00000000-0000-0000-0000-000000007009', 'RIB-TANK-BLK-S', 'Black / S', 2200, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008011', '00000000-0000-0000-0000-000000007010', 'TRAVEL-BACKPACK-CHARCOAL', 'Charcoal', 7800, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008012', '00000000-0000-0000-0000-000000007011', 'ANC-HEADPHONES-GRAPHITE', 'Graphite', 12900, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008013', '00000000-0000-0000-0000-000000007012', 'CHARGE-STAND-SILVER', 'Silver', 3600, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008014', '00000000-0000-0000-0000-000000007013', 'LIP-TINT-SET-3', 'Set of 3', 3200, 'USD', TRUE),
  ('00000000-0000-0000-0000-000000008015', '00000000-0000-0000-0000-000000007014', 'ACACIA-BOARD-M', 'Medium', 3400, 'USD', TRUE)
ON CONFLICT (id) DO UPDATE
SET product_id = EXCLUDED.product_id,
    sku = EXCLUDED.sku,
    name = EXCLUDED.name,
    price_cents = EXCLUDED.price_cents,
    currency = EXCLUDED.currency,
    is_active = EXCLUDED.is_active;

INSERT INTO inventory_items (product_id, variant_id, quantity_available, quantity_reserved, low_stock_threshold)
VALUES
  ('00000000-0000-0000-0000-000000007001', '00000000-0000-0000-0000-000000008001', 74, 0, 10),
  ('00000000-0000-0000-0000-000000007001', '00000000-0000-0000-0000-000000008002', 92, 0, 10),
  ('00000000-0000-0000-0000-000000007002', '00000000-0000-0000-0000-000000008003', 41, 0, 6),
  ('00000000-0000-0000-0000-000000007003', '00000000-0000-0000-0000-000000008004', 65, 0, 8),
  ('00000000-0000-0000-0000-000000007004', '00000000-0000-0000-0000-000000008005', 24, 0, 5),
  ('00000000-0000-0000-0000-000000007005', '00000000-0000-0000-0000-000000008006', 33, 0, 5),
  ('00000000-0000-0000-0000-000000007006', '00000000-0000-0000-0000-000000008007', 59, 0, 8),
  ('00000000-0000-0000-0000-000000007007', '00000000-0000-0000-0000-000000008008', 120, 0, 12),
  ('00000000-0000-0000-0000-000000007008', '00000000-0000-0000-0000-000000008009', 145, 0, 15),
  ('00000000-0000-0000-0000-000000007009', '00000000-0000-0000-0000-000000008010', 88, 0, 10),
  ('00000000-0000-0000-0000-000000007010', '00000000-0000-0000-0000-000000008011', 37, 0, 6),
  ('00000000-0000-0000-0000-000000007011', '00000000-0000-0000-0000-000000008012', 28, 0, 5),
  ('00000000-0000-0000-0000-000000007012', '00000000-0000-0000-0000-000000008013', 73, 0, 8),
  ('00000000-0000-0000-0000-000000007013', '00000000-0000-0000-0000-000000008014', 104, 0, 12),
  ('00000000-0000-0000-0000-000000007014', '00000000-0000-0000-0000-000000008015', 44, 0, 6)
ON CONFLICT (variant_id) DO UPDATE
SET product_id = EXCLUDED.product_id,
    quantity_available = EXCLUDED.quantity_available,
    low_stock_threshold = EXCLUDED.low_stock_threshold;

INSERT INTO product_images (id, product_id, url, alt_text, sort_order, is_primary)
VALUES
  ('00000000-0000-0000-0000-000000009001', '00000000-0000-0000-0000-000000007001', 'https://images.unsplash.com/photo-1521572163474-6864f9cf17ab?auto=format&fit=crop&w=1200&q=80', 'Organic Cotton Relaxed Tee', 10, TRUE),
  ('00000000-0000-0000-0000-000000009002', '00000000-0000-0000-0000-000000007002', 'https://images.unsplash.com/photo-1544441893-675973e31985?auto=format&fit=crop&w=1200&q=80', 'Linen Overshirt', 10, TRUE),
  ('00000000-0000-0000-0000-000000009003', '00000000-0000-0000-0000-000000007003', 'https://images.unsplash.com/photo-1590874103328-eac38a683ce7?auto=format&fit=crop&w=1200&q=80', 'Everyday Canvas Tote', 10, TRUE),
  ('00000000-0000-0000-0000-000000009004', '00000000-0000-0000-0000-000000007004', 'https://images.unsplash.com/photo-1610701596007-11502861dcfa?auto=format&fit=crop&w=1200&q=80', 'Stoneware Dinner Set', 10, TRUE),
  ('00000000-0000-0000-0000-000000009005', '00000000-0000-0000-0000-000000007005', 'https://images.unsplash.com/photo-1608354580875-30bd4168b351?auto=format&fit=crop&w=1200&q=80', 'Matte Pour-Over Kettle', 10, TRUE),
  ('00000000-0000-0000-0000-000000009006', '00000000-0000-0000-0000-000000007006', 'https://images.unsplash.com/photo-1583845112203-454c095cb7c8?auto=format&fit=crop&w=1200&q=80', 'Textured Cotton Throw', 10, TRUE),
  ('00000000-0000-0000-0000-000000009007', '00000000-0000-0000-0000-000000007007', 'https://images.unsplash.com/photo-1620916566398-39f1143ab7be?auto=format&fit=crop&w=1200&q=80', 'Hydrating Serum Duo', 10, TRUE),
  ('00000000-0000-0000-0000-000000009008', '00000000-0000-0000-0000-000000007008', 'https://images.unsplash.com/photo-1556228578-8c89e6adf883?auto=format&fit=crop&w=1200&q=80', 'Mineral Sunscreen SPF 50', 10, TRUE),
  ('00000000-0000-0000-0000-000000009009', '00000000-0000-0000-0000-000000007009', 'https://images.unsplash.com/photo-1506629905607-d9c297d74d9d?auto=format&fit=crop&w=1200&q=80', 'Slim Rib Tank', 10, TRUE),
  ('00000000-0000-0000-0000-000000009010', '00000000-0000-0000-0000-000000007010', 'https://images.unsplash.com/photo-1553062407-98eeb64c6a62?auto=format&fit=crop&w=1200&q=80', 'Compact Travel Backpack', 10, TRUE),
  ('00000000-0000-0000-0000-000000009011', '00000000-0000-0000-0000-000000007011', 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=1200&q=80', 'Noise Cancelling Headphones', 10, TRUE),
  ('00000000-0000-0000-0000-000000009012', '00000000-0000-0000-0000-000000007012', 'https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?auto=format&fit=crop&w=1200&q=80', 'Wireless Charging Stand', 10, TRUE),
  ('00000000-0000-0000-0000-000000009013', '00000000-0000-0000-0000-000000007013', 'https://images.unsplash.com/photo-1599305090598-fe179d501227?auto=format&fit=crop&w=1200&q=80', 'Clean Lip Tint Set', 10, TRUE),
  ('00000000-0000-0000-0000-000000009014', '00000000-0000-0000-0000-000000007014', 'https://images.unsplash.com/photo-1593618998160-e34014e67546?auto=format&fit=crop&w=1200&q=80', 'Acacia Cutting Board', 10, TRUE)
ON CONFLICT (id) DO UPDATE
SET url = EXCLUDED.url,
    alt_text = EXCLUDED.alt_text,
    sort_order = EXCLUDED.sort_order,
    is_primary = EXCLUDED.is_primary;

INSERT INTO product_reviews (id, product_id, user_id, rating, title, body, is_verified_purchase)
VALUES
  ('00000000-0000-0000-0000-000000010001', '00000000-0000-0000-0000-000000007001', '00000000-0000-0000-0000-000000001003', 5, 'Perfect weight', 'Soft, structured, and easy to wear twice a week.', TRUE),
  ('00000000-0000-0000-0000-000000010002', '00000000-0000-0000-0000-000000007002', '00000000-0000-0000-0000-000000001003', 5, 'Looks expensive', 'Great layering piece and the linen feels substantial.', TRUE),
  ('00000000-0000-0000-0000-000000010003', '00000000-0000-0000-0000-000000007004', '00000000-0000-0000-0000-000000001003', 5, 'Beautiful set', 'The glaze is even nicer in person and packed safely.', TRUE),
  ('00000000-0000-0000-0000-000000010004', '00000000-0000-0000-0000-000000007007', '00000000-0000-0000-0000-000000001003', 5, 'Light and effective', 'Absorbs fast and works well under sunscreen.', TRUE),
  ('00000000-0000-0000-0000-000000010005', '00000000-0000-0000-0000-000000007011', '00000000-0000-0000-0000-000000001003', 4, 'Strong battery', 'Comfortable for work calls and commuting.', TRUE)
ON CONFLICT (id) DO UPDATE
SET rating = EXCLUDED.rating,
    title = EXCLUDED.title,
    body = EXCLUDED.body,
    is_verified_purchase = EXCLUDED.is_verified_purchase;

INSERT INTO cms_banners (id, title, image_url, link_url, placement, sort_order, status, starts_at)
VALUES
  ('00000000-0000-0000-0000-000000011001', 'New season essentials', 'https://images.unsplash.com/photo-1496747611176-843222e1e57c?auto=format&fit=crop&w=1800&q=80', '/public/products?sort=newest', 'home', 10, 'published', now()),
  ('00000000-0000-0000-0000-000000011002', 'Home objects that feel considered', 'https://images.unsplash.com/photo-1513161455079-7dc1de15ef3e?auto=format&fit=crop&w=1800&q=80', '/public/categories/00000000-0000-0000-0000-000000002003/products', 'home', 20, 'published', now()),
  ('00000000-0000-0000-0000-000000011003', 'Clean beauty for every day', 'https://images.unsplash.com/photo-1596462502278-27bfdc403348?auto=format&fit=crop&w=1800&q=80', '/public/categories/00000000-0000-0000-0000-000000002004/products', 'home', 30, 'published', now())
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
  ('home.featured_categories', '["fashion","electronics","home-living","beauty","bags-travel","kitchen"]', 'Homepage featured category slugs.', TRUE),
  ('home.featured_products', '["organic-cotton-relaxed-tee","stoneware-dinner-set-cream","hydrating-serum-duo","noise-cancelling-headphones","compact-travel-backpack","matte-pour-over-kettle"]', 'Homepage featured product slugs.', TRUE),
  ('home.layout', '{"sections":["banners","featured_categories","featured_products"]}', 'Homepage section order.', TRUE)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    description = EXCLUDED.description,
    is_public = EXCLUDED.is_public;
