-- +goose Up
-- +goose StatementBegin
ALTER TABLE cart_items
  ADD COLUMN IF NOT EXISTS selected BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE carts
  ADD COLUMN IF NOT EXISTS voucher_id UUID REFERENCES vouchers(id) ON DELETE SET NULL;

ALTER TABLE checkout_sessions
  ADD COLUMN IF NOT EXISTS voucher_id UUID REFERENCES vouchers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_cart_items_selected ON cart_items(cart_id, selected);
CREATE INDEX IF NOT EXISTS idx_carts_voucher_id ON carts(voucher_id) WHERE voucher_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_checkout_sessions_voucher_id ON checkout_sessions(voucher_id) WHERE voucher_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS order_invoices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
  invoice_number VARCHAR(80) NOT NULL UNIQUE,
  status VARCHAR(40) NOT NULL DEFAULT 'generated',
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_order_invoices_order ON order_invoices(order_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_invoices;
DROP INDEX IF EXISTS idx_checkout_sessions_voucher_id;
DROP INDEX IF EXISTS idx_carts_voucher_id;
DROP INDEX IF EXISTS idx_cart_items_selected;
ALTER TABLE checkout_sessions DROP COLUMN IF EXISTS voucher_id;
ALTER TABLE carts DROP COLUMN IF EXISTS voucher_id;
ALTER TABLE cart_items DROP COLUMN IF EXISTS selected;
-- +goose StatementEnd
