-- +goose Up
-- +goose StatementBegin
CREATE TABLE vouchers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code VARCHAR(80) NOT NULL UNIQUE,
  title VARCHAR(180) NOT NULL,
  description TEXT,
  discount_type VARCHAR(40) NOT NULL CHECK (discount_type IN ('fixed', 'percentage', 'free_shipping')),
  discount_value BIGINT NOT NULL CHECK (discount_value >= 0),
  min_order_cents BIGINT NOT NULL DEFAULT 0 CHECK (min_order_cents >= 0),
  max_discount_cents BIGINT CHECK (max_discount_cents IS NULL OR max_discount_cents >= 0),
  usage_limit INT CHECK (usage_limit IS NULL OR usage_limit >= 0),
  usage_count INT NOT NULL DEFAULT 0 CHECK (usage_count >= 0),
  starts_at TIMESTAMPTZ NOT NULL,
  ends_at TIMESTAMPTZ NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (ends_at > starts_at)
);

CREATE INDEX idx_vouchers_code ON vouchers(code);

CREATE INDEX idx_vouchers_active_dates ON vouchers(is_active, starts_at, ends_at);

CREATE TABLE voucher_redemptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  voucher_id UUID NOT NULL REFERENCES vouchers(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
  discount_cents BIGINT NOT NULL DEFAULT 0 CHECK (discount_cents >= 0),
  redeemed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (voucher_id, user_id, order_id)
);

CREATE TABLE product_reviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  order_item_id UUID REFERENCES order_items(id) ON DELETE SET NULL,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
  title VARCHAR(180),
  body TEXT,
  is_verified_purchase BOOLEAN NOT NULL DEFAULT FALSE,
  is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (order_item_id, user_id)
);

CREATE INDEX idx_product_reviews_product ON product_reviews(product_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_product_reviews_user ON product_reviews(user_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_product_reviews_updated_at BEFORE UPDATE ON product_reviews FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE product_review_images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  review_id UUID NOT NULL REFERENCES product_reviews(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE review_votes (
  review_id UUID NOT NULL REFERENCES product_reviews(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  vote INT NOT NULL CHECK (vote IN (-1, 1)),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (review_id, user_id)
);

CREATE TABLE review_reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  review_id UUID NOT NULL REFERENCES product_reviews(id) ON DELETE CASCADE,
  reporter_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  reason TEXT NOT NULL,
  status moderation_status NOT NULL DEFAULT 'open',
  resolved_by UUID REFERENCES users(id) ON DELETE SET NULL,
  resolved_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_review_reports_status ON review_reports(status, created_at);

CREATE TABLE wishlists (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE wishlist_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wishlist_id UUID NOT NULL REFERENCES wishlists(id) ON DELETE CASCADE,
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (wishlist_id, product_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wishlist_items;
DROP TABLE IF EXISTS wishlists;
DROP TABLE IF EXISTS review_reports;
DROP TABLE IF EXISTS review_votes;
DROP TABLE IF EXISTS product_review_images;
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS voucher_redemptions;
DROP TABLE IF EXISTS vouchers;
-- +goose StatementEnd
