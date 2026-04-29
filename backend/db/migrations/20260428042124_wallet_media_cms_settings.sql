-- +goose Up
-- +goose StatementBegin
CREATE TABLE seller_wallets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seller_id UUID NOT NULL UNIQUE REFERENCES sellers(id) ON DELETE CASCADE,
  balance_cents BIGINT NOT NULL DEFAULT 0 CHECK (balance_cents >= 0),
  pending_balance_cents BIGINT NOT NULL DEFAULT 0 CHECK (pending_balance_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE seller_wallet_transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seller_wallet_id UUID NOT NULL REFERENCES seller_wallets(id) ON DELETE CASCADE,
  type wallet_tx_type NOT NULL,
  amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
  balance_before_cents BIGINT NOT NULL CHECK (balance_before_cents >= 0),
  balance_after_cents BIGINT NOT NULL CHECK (balance_after_cents >= 0),
  reference_type VARCHAR(100),
  reference_id UUID,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_seller_wallet_tx_wallet ON seller_wallet_transactions(seller_wallet_id);

CREATE TABLE seller_payouts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seller_id UUID NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
  amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  status VARCHAR(50) NOT NULL DEFAULT 'pending',
  provider VARCHAR(80),
  provider_payout_id VARCHAR(255),
  requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  processed_at TIMESTAMPTZ
);

CREATE INDEX idx_seller_payouts_seller ON seller_payouts(seller_id);

CREATE TABLE media_files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  url TEXT NOT NULL,
  bucket VARCHAR(120),
  object_key TEXT,
  mime_type VARCHAR(120),
  size_bytes BIGINT CHECK (size_bytes IS NULL OR size_bytes >= 0),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_media_owner ON media_files(owner_user_id);

CREATE TABLE cms_pages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug VARCHAR(220) NOT NULL UNIQUE,
  title VARCHAR(220) NOT NULL,
  body TEXT NOT NULL,
  status cms_status NOT NULL DEFAULT 'draft',
  published_at TIMESTAMPTZ,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cms_pages_status ON cms_pages(status, published_at);

CREATE TRIGGER trg_cms_pages_updated_at BEFORE UPDATE ON cms_pages FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE cms_banners (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(220) NOT NULL,
  image_url TEXT NOT NULL,
  link_url TEXT,
  placement VARCHAR(100) NOT NULL DEFAULT 'home',
  sort_order INT NOT NULL DEFAULT 0,
  status cms_status NOT NULL DEFAULT 'draft',
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX idx_cms_banners_placement ON cms_banners(placement, status, sort_order);

CREATE TRIGGER trg_cms_banners_updated_at BEFORE UPDATE ON cms_banners FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE cms_faqs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question TEXT NOT NULL,
  answer TEXT NOT NULL,
  category VARCHAR(100),
  sort_order INT NOT NULL DEFAULT 0,
  status cms_status NOT NULL DEFAULT 'draft',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cms_faqs_category ON cms_faqs(category, status, sort_order);

CREATE TRIGGER trg_cms_faqs_updated_at BEFORE UPDATE ON cms_faqs FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE cms_announcements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(220) NOT NULL,
  body TEXT NOT NULL,
  audience VARCHAR(80) NOT NULL DEFAULT 'all',
  status cms_status NOT NULL DEFAULT 'draft',
  starts_at TIMESTAMPTZ,
  ends_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (ends_at IS NULL OR starts_at IS NULL OR ends_at > starts_at)
);

CREATE INDEX idx_cms_announcements_audience ON cms_announcements(audience, status, starts_at);

CREATE TRIGGER trg_cms_announcements_updated_at BEFORE UPDATE ON cms_announcements FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE feature_flags (
  key VARCHAR(150) PRIMARY KEY,
  description TEXT,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  rules JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_feature_flags_updated_at BEFORE UPDATE ON feature_flags FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE platform_settings (
  key VARCHAR(150) PRIMARY KEY,
  value JSONB NOT NULL,
  description TEXT,
  is_public BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_platform_settings_updated_at BEFORE UPDATE ON platform_settings FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform_settings;
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS cms_announcements;
DROP TABLE IF EXISTS cms_faqs;
DROP TABLE IF EXISTS cms_banners;
DROP TABLE IF EXISTS cms_pages;
DROP TABLE IF EXISTS media_files;
DROP TABLE IF EXISTS seller_payouts;
DROP TABLE IF EXISTS seller_wallet_transactions;
DROP TABLE IF EXISTS seller_wallets;
-- +goose StatementEnd
