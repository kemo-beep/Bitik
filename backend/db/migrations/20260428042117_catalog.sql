-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
  name VARCHAR(180) NOT NULL,
  slug VARCHAR(220) NOT NULL UNIQUE,
  image_url TEXT,
  sort_order INT NOT NULL DEFAULT 0,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_categories_parent_id ON categories(parent_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_categories_active_sort ON categories(is_active, sort_order) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE brands (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(180) NOT NULL UNIQUE,
  slug VARCHAR(220) NOT NULL UNIQUE,
  logo_url TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_brands_active ON brands(is_active) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_brands_updated_at BEFORE UPDATE ON brands FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE products (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seller_id UUID NOT NULL REFERENCES sellers(id) ON DELETE CASCADE,
  category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
  brand_id UUID REFERENCES brands(id) ON DELETE SET NULL,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(280) NOT NULL UNIQUE,
  description TEXT,
  status product_status NOT NULL DEFAULT 'draft',
  min_price_cents BIGINT NOT NULL DEFAULT 0 CHECK (min_price_cents >= 0),
  max_price_cents BIGINT NOT NULL DEFAULT 0 CHECK (max_price_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  total_sold BIGINT NOT NULL DEFAULT 0 CHECK (total_sold >= 0),
  rating NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
  review_count BIGINT NOT NULL DEFAULT 0 CHECK (review_count >= 0),
  search_vector TSVECTOR,
  published_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (max_price_cents >= min_price_cents)
);

CREATE INDEX idx_products_seller_id ON products(seller_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_products_category_id ON products(category_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_products_status ON products(status) WHERE deleted_at IS NULL;

CREATE INDEX idx_products_created_at ON products(created_at);

CREATE INDEX idx_products_search_vector ON products USING GIN(search_vector);

CREATE TRIGGER trg_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE product_images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  alt_text VARCHAR(255),
  sort_order INT NOT NULL DEFAULT 0,
  is_primary BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);

CREATE UNIQUE INDEX idx_product_images_one_primary ON product_images(product_id) WHERE is_primary = TRUE;

CREATE TABLE product_variants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  sku VARCHAR(120) NOT NULL UNIQUE,
  name VARCHAR(255),
  price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
  compare_at_price_cents BIGINT CHECK (compare_at_price_cents IS NULL OR compare_at_price_cents >= 0),
  currency VARCHAR(10) NOT NULL DEFAULT 'USD',
  weight_grams INT CHECK (weight_grams IS NULL OR weight_grams >= 0),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_product_variants_product_id ON product_variants(product_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_product_variants_updated_at BEFORE UPDATE ON product_variants FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE variant_options (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  UNIQUE (product_id, name)
);

CREATE TABLE variant_option_values (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  option_id UUID NOT NULL REFERENCES variant_options(id) ON DELETE CASCADE,
  value VARCHAR(120) NOT NULL,
  sort_order INT NOT NULL DEFAULT 0,
  UNIQUE (option_id, value)
);

CREATE TABLE product_variant_option_values (
  variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE CASCADE,
  option_value_id UUID NOT NULL REFERENCES variant_option_values(id) ON DELETE CASCADE,
  PRIMARY KEY (variant_id, option_value_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_variant_option_values;
DROP TABLE IF EXISTS variant_option_values;
DROP TABLE IF EXISTS variant_options;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS brands;
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
