-- +goose Up
-- +goose StatementBegin
ALTER TABLE vouchers
  ADD COLUMN seller_id UUID REFERENCES sellers(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_vouchers_seller_id ON vouchers(seller_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_vouchers_seller_id;
ALTER TABLE vouchers DROP COLUMN IF EXISTS seller_id;
-- +goose StatementEnd
