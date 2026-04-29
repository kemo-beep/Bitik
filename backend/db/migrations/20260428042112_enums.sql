-- +goose Up
-- +goose StatementBegin
-- Idempotent: safe when types already exist (e.g. DB restored without goose_db_version).
DO $create$ BEGIN
  CREATE TYPE user_status AS ENUM ('active', 'inactive', 'banned', 'deleted');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE seller_status AS ENUM ('pending', 'active', 'suspended', 'rejected');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE seller_application_status AS ENUM ('draft', 'submitted', 'in_review', 'approved', 'rejected', 'cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE product_status AS ENUM ('draft', 'active', 'inactive', 'banned', 'deleted');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE order_status AS ENUM ('pending_payment', 'paid', 'processing', 'shipped', 'delivered', 'completed', 'cancelled', 'refunded', 'disputed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE payment_status AS ENUM ('pending', 'authorized', 'paid', 'failed', 'refunded', 'cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE shipment_status AS ENUM ('pending', 'packed', 'shipped', 'in_transit', 'delivered', 'failed', 'returned');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE wallet_tx_type AS ENUM ('credit', 'debit');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE inventory_movement_type AS ENUM ('stock_in', 'stock_out', 'reserve', 'release', 'adjustment', 'return');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE refund_status AS ENUM ('requested', 'approved', 'rejected', 'processing', 'refunded', 'cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE return_status AS ENUM ('requested', 'approved', 'rejected', 'received', 'refunded', 'cancelled');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE moderation_status AS ENUM ('open', 'under_review', 'resolved', 'dismissed');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;

DO $create$ BEGIN
  CREATE TYPE cms_status AS ENUM ('draft', 'published', 'archived');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $create$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS cms_status;
DROP TYPE IF EXISTS moderation_status;
DROP TYPE IF EXISTS return_status;
DROP TYPE IF EXISTS refund_status;
DROP TYPE IF EXISTS inventory_movement_type;
DROP TYPE IF EXISTS wallet_tx_type;
DROP TYPE IF EXISTS shipment_status;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS product_status;
DROP TYPE IF EXISTS seller_application_status;
DROP TYPE IF EXISTS seller_status;
DROP TYPE IF EXISTS user_status;
-- +goose StatementEnd
