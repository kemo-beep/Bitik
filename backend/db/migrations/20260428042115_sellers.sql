-- +goose Up
-- +goose StatementBegin
CREATE TABLE seller_applications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  shop_name VARCHAR(180) NOT NULL,
  slug VARCHAR(220) NOT NULL,
  business_type VARCHAR(80),
  country VARCHAR(80),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  status seller_application_status NOT NULL DEFAULT 'draft',
  submitted_at TIMESTAMPTZ,
  reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at TIMESTAMPTZ,
  rejection_reason TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_seller_applications_user ON seller_applications(user_id);

CREATE INDEX idx_seller_applications_status ON seller_applications(status, submitted_at);

CREATE UNIQUE INDEX idx_seller_applications_active_slug ON seller_applications(slug) WHERE status <> 'cancelled';

CREATE TRIGGER trg_seller_applications_updated_at BEFORE UPDATE ON seller_applications FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE sellers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  application_id UUID UNIQUE REFERENCES seller_applications(id) ON DELETE SET NULL,
  shop_name VARCHAR(180) NOT NULL,
  slug VARCHAR(220) NOT NULL UNIQUE,
  description TEXT,
  logo_url TEXT,
  banner_url TEXT,
  status seller_status NOT NULL DEFAULT 'pending',
  rating NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
  total_sales BIGINT NOT NULL DEFAULT 0 CHECK (total_sales >= 0),
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sellers_status ON sellers(status) WHERE deleted_at IS NULL;

CREATE INDEX idx_sellers_created_at ON sellers(created_at);

CREATE TRIGGER trg_sellers_updated_at BEFORE UPDATE ON sellers FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE seller_documents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seller_application_id UUID REFERENCES seller_applications(id) ON DELETE CASCADE,
  seller_id UUID REFERENCES sellers(id) ON DELETE CASCADE,
  document_type VARCHAR(100) NOT NULL,
  file_url TEXT NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'pending',
  reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at TIMESTAMPTZ,
  rejection_reason TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (seller_application_id IS NOT NULL OR seller_id IS NOT NULL)
);

CREATE INDEX idx_seller_documents_application ON seller_documents(seller_application_id);

CREATE INDEX idx_seller_documents_seller ON seller_documents(seller_id);

CREATE TABLE seller_bank_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seller_id UUID NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
  bank_name VARCHAR(150) NOT NULL,
  account_name VARCHAR(150) NOT NULL,
  account_number_masked VARCHAR(80) NOT NULL,
  country VARCHAR(80),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_seller_bank_accounts_seller ON seller_bank_accounts(seller_id);

CREATE UNIQUE INDEX idx_seller_bank_accounts_one_default ON seller_bank_accounts(seller_id) WHERE is_default = TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS seller_bank_accounts;
DROP TABLE IF EXISTS seller_documents;
DROP TABLE IF EXISTS sellers;
DROP TABLE IF EXISTS seller_applications;
-- +goose StatementEnd
