-- +goose Up
-- +goose StatementBegin
CREATE TABLE shipping_providers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(150) NOT NULL,
  code VARCHAR(80) NOT NULL UNIQUE,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_shipping_providers_updated_at BEFORE UPDATE ON shipping_providers FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE shipments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  seller_id UUID NOT NULL REFERENCES sellers(id) ON DELETE RESTRICT,
  provider_id UUID REFERENCES shipping_providers(id) ON DELETE SET NULL,
  tracking_number VARCHAR(150),
  label_url TEXT,
  provider_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  status shipment_status NOT NULL DEFAULT 'pending',
  shipped_at TIMESTAMPTZ,
  delivered_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_shipments_order_id ON shipments(order_id);

CREATE INDEX idx_shipments_seller_id ON shipments(seller_id);

CREATE INDEX idx_shipments_tracking_number ON shipments(tracking_number);

CREATE INDEX idx_shipments_status ON shipments(status, created_at);

CREATE TRIGGER trg_shipments_updated_at BEFORE UPDATE ON shipments FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE shipment_tracking_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
  status VARCHAR(120) NOT NULL,
  location VARCHAR(255),
  message TEXT,
  event_time TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tracking_events_shipment ON shipment_tracking_events(shipment_id);

CREATE TABLE shipment_labels (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
  label_url TEXT NOT NULL,
  format VARCHAR(40) NOT NULL DEFAULT 'pdf',
  generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_shipment_labels_shipment ON shipment_labels(shipment_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shipment_labels;
DROP TABLE IF EXISTS shipment_tracking_events;
DROP TABLE IF EXISTS shipments;
DROP TABLE IF EXISTS shipping_providers;
-- +goose StatementEnd
