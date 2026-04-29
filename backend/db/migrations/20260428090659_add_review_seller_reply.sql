-- +goose Up
-- +goose StatementBegin
ALTER TABLE product_reviews
  ADD COLUMN seller_reply TEXT,
  ADD COLUMN seller_reply_at TIMESTAMPTZ,
  ADD COLUMN seller_reply_by UUID REFERENCES users(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE product_reviews
  DROP COLUMN IF EXISTS seller_reply_by,
  DROP COLUMN IF EXISTS seller_reply_at,
  DROP COLUMN IF EXISTS seller_reply;
-- +goose StatementEnd
