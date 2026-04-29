-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_methods (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider VARCHAR(80) NOT NULL,
  type VARCHAR(80) NOT NULL,
  display_name VARCHAR(150),
  token_reference TEXT,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payment_methods_user ON payment_methods(user_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX idx_payment_methods_one_default ON payment_methods(user_id) WHERE is_default = TRUE AND deleted_at IS NULL;

CREATE TRIGGER trg_payment_methods_updated_at BEFORE UPDATE ON payment_methods FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  payment_method_id UUID REFERENCES payment_methods(id) ON DELETE SET NULL,
  provider VARCHAR(80) NOT NULL,
  provider_payment_id VARCHAR(255),
  status payment_status NOT NULL DEFAULT 'pending',
  amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  idempotency_key VARCHAR(255) UNIQUE,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  paid_at TIMESTAMPTZ,
  failed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_payments_order_id ON payments(order_id);

CREATE INDEX idx_payments_provider_payment_id ON payments(provider, provider_payment_id);

CREATE INDEX idx_payments_status ON payments(status);

CREATE TRIGGER trg_payments_updated_at BEFORE UPDATE ON payments FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE manual_wave_payment_approvals (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_id UUID NOT NULL UNIQUE REFERENCES payments(id) ON DELETE CASCADE,
  reference VARCHAR(160) NOT NULL,
  sender_phone VARCHAR(32),
  amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
  approved_at TIMESTAMPTZ,
  rejected_by UUID REFERENCES users(id) ON DELETE SET NULL,
  rejected_at TIMESTAMPTZ,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_manual_wave_payment_reference ON manual_wave_payment_approvals(reference);

CREATE TABLE pod_payment_captures (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_id UUID NOT NULL UNIQUE REFERENCES payments(id) ON DELETE CASCADE,
  captured_by UUID REFERENCES users(id) ON DELETE SET NULL,
  amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  captured_at TIMESTAMPTZ NOT NULL,
  note TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payment_webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  provider VARCHAR(80) NOT NULL,
  event_id VARCHAR(255) NOT NULL,
  event_type VARCHAR(150),
  payload JSONB NOT NULL,
  processed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (provider, event_id)
);

CREATE INDEX idx_payment_webhook_events_processed ON payment_webhook_events(provider, processed_at);

CREATE TABLE refunds (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  payment_id UUID REFERENCES payments(id) ON DELETE SET NULL,
  requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
  status refund_status NOT NULL DEFAULT 'requested',
  reason TEXT,
  amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at TIMESTAMPTZ,
  processed_at TIMESTAMPTZ,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refunds_order ON refunds(order_id);

CREATE INDEX idx_refunds_status ON refunds(status, created_at);

CREATE TRIGGER trg_refunds_updated_at BEFORE UPDATE ON refunds FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE return_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  order_item_id UUID REFERENCES order_items(id) ON DELETE SET NULL,
  requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
  status return_status NOT NULL DEFAULT 'requested',
  reason TEXT,
  quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
  reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at TIMESTAMPTZ,
  received_at TIMESTAMPTZ,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_return_requests_order ON return_requests(order_id);

CREATE INDEX idx_return_requests_status ON return_requests(status, created_at);

CREATE TRIGGER trg_return_requests_updated_at BEFORE UPDATE ON return_requests FOR EACH ROW EXECUTE FUNCTION set_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS return_requests;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_webhook_events;
DROP TABLE IF EXISTS pod_payment_captures;
DROP TABLE IF EXISTS manual_wave_payment_approvals;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS payment_methods;
-- +goose StatementEnd
