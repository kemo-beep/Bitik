-- +goose Up
-- +goose StatementBegin
CREATE TABLE inventory_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  variant_id UUID NOT NULL UNIQUE REFERENCES product_variants(id) ON DELETE CASCADE,
  quantity_available INT NOT NULL DEFAULT 0 CHECK (quantity_available >= 0),
  quantity_reserved INT NOT NULL DEFAULT 0 CHECK (quantity_reserved >= 0),
  low_stock_threshold INT NOT NULL DEFAULT 5 CHECK (low_stock_threshold >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inventory_product_id ON inventory_items(product_id);

CREATE INDEX idx_inventory_low_stock ON inventory_items(product_id) WHERE quantity_available <= low_stock_threshold;

CREATE TRIGGER trg_inventory_items_updated_at BEFORE UPDATE ON inventory_items FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE inventory_movements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inventory_item_id UUID NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
  movement_type inventory_movement_type NOT NULL,
  quantity INT NOT NULL CHECK (quantity > 0),
  reason TEXT,
  reference_type VARCHAR(100),
  reference_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_inventory_movements_item ON inventory_movements(inventory_item_id);

CREATE INDEX idx_inventory_movements_reference ON inventory_movements(reference_type, reference_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory_items;
-- +goose StatementEnd
