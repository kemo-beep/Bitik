# Public Catalog And Media

Phase 3 adds two backend modules:

- `internal/catalogsvc`: public read APIs for homepage content, banners, categories, brands, sellers, products, variants, reviews, and related products.
- `internal/mediasvc`: seller/admin media upload APIs backed by the S3-compatible MinIO client, usable with local MinIO and production R2-style storage.

## Catalog

Public routes are mounted under `/api/v1/public`.

Product lists support:

- filters: `q`, `category_id`, `brand_id`, `seller_id`, `min_price_cents`, `max_price_cents`
- sorting: `newest`, `price_asc`, `price_desc`, `popular`, `rating`
- pagination: `page`, `per_page` with a max page size of `100`

Public product queries only return active products owned by active, non-deleted sellers. Invalid UUID or price filters return `400` instead of silently broadening the result set.

The Phase 3 migration adds public-list indexes and a `products.search_vector` trigger so text search stays current on product name/description changes.

## Media

Media routes are mounted under `/api/v1/media` and require JWT authentication plus an active account. Upload creation is seller/admin scoped.

Supported upload flows:

- `POST /media/upload`: multipart upload through the API.
- `POST /media/upload/presigned-url`: reserve pending metadata and return a short-lived direct-upload URL.
- `POST /media/upload/presigned-complete`: fetch the uploaded object, validate bytes, scan through the hook, and mark metadata ready.
- `GET /media/files`: list the current user's files.
- `GET /media/files/{file_id}`: read file metadata.
- `DELETE /media/files/{file_id}`: delete owned media metadata and the object when storage is available.

Validation currently enforces byte-sniffed MIME type, extension, size, image dimensions, ownership checks, and a `MalwareScanner` hook. The `ImageProcessor` hook is called after successful API uploads and presigned completion so a worker can later add thumbnails, optimization, or moderation jobs.

Storage uses the existing `storage` config (`endpoint`, `bucket`, keys, SSL). MinIO is the local provider; R2 and other S3-compatible providers use the same adapter.
