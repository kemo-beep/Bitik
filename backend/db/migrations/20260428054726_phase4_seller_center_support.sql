-- +goose Up
-- +goose StatementBegin
ALTER TABLE sellers
  ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE products
  ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(50) NOT NULL DEFAULT 'approved',
  ADD COLUMN IF NOT EXISTS moderation_reason TEXT,
  ADD COLUMN IF NOT EXISTS moderated_by UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS moderated_at TIMESTAMPTZ;

ALTER TABLE inventory_movements
  ADD COLUMN IF NOT EXISTS actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS before_available INT,
  ADD COLUMN IF NOT EXISTS after_available INT,
  ADD COLUMN IF NOT EXISTS before_reserved INT,
  ADD COLUMN IF NOT EXISTS after_reserved INT;

CREATE INDEX IF NOT EXISTS idx_products_seller_status_moderation
ON products (seller_id, status, moderation_status)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_inventory_movements_created
ON inventory_movements (inventory_item_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_seller_documents_created
ON seller_documents (seller_application_id, seller_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_seller_documents_created;
DROP INDEX IF EXISTS idx_inventory_movements_created;
DROP INDEX IF EXISTS idx_products_seller_status_moderation;

ALTER TABLE inventory_movements
  DROP COLUMN IF EXISTS after_reserved,
  DROP COLUMN IF EXISTS before_reserved,
  DROP COLUMN IF EXISTS after_available,
  DROP COLUMN IF EXISTS before_available,
  DROP COLUMN IF EXISTS actor_user_id;

ALTER TABLE products
  DROP COLUMN IF EXISTS moderated_at,
  DROP COLUMN IF EXISTS moderated_by,
  DROP COLUMN IF EXISTS moderation_reason,
  DROP COLUMN IF EXISTS moderation_status;

ALTER TABLE sellers
  DROP COLUMN IF EXISTS settings;
-- +goose StatementEnd
