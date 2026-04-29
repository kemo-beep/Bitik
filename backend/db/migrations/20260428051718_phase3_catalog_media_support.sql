-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_products_public_category_sort
ON products (category_id, status, published_at DESC, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_public_brand_sort
ON products (brand_id, status, published_at DESC, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_public_seller_sort
ON products (seller_id, status, published_at DESC, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_public_price
ON products (status, min_price_cents, max_price_cents)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_product_reviews_public_product_created
ON product_reviews (product_id, created_at DESC)
WHERE deleted_at IS NULL AND is_hidden = FALSE;

CREATE INDEX IF NOT EXISTS idx_cms_banners_public_window
ON cms_banners (placement, status, sort_order, starts_at, ends_at);

CREATE INDEX IF NOT EXISTS idx_media_files_owner_created
ON media_files (owner_user_id, created_at DESC);

CREATE OR REPLACE FUNCTION set_product_search_vector()
RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', COALESCE(NEW.name, '')), 'A') ||
    setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_products_search_vector ON products;
CREATE TRIGGER trg_products_search_vector
BEFORE INSERT OR UPDATE OF name, description
ON products
FOR EACH ROW
EXECUTE FUNCTION set_product_search_vector();

UPDATE products
SET search_vector =
  setweight(to_tsvector('simple', COALESCE(name, '')), 'A') ||
  setweight(to_tsvector('simple', COALESCE(description, '')), 'B')
WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_products_search_vector ON products;
DROP FUNCTION IF EXISTS set_product_search_vector();
DROP INDEX IF EXISTS idx_media_files_owner_created;
DROP INDEX IF EXISTS idx_cms_banners_public_window;
DROP INDEX IF EXISTS idx_product_reviews_public_product_created;
DROP INDEX IF EXISTS idx_products_public_price;
DROP INDEX IF EXISTS idx_products_public_seller_sort;
DROP INDEX IF EXISTS idx_products_public_brand_sort;
DROP INDEX IF EXISTS idx_products_public_category_sort;
-- +goose StatementEnd
