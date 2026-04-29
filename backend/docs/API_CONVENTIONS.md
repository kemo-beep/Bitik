# Bitik API Conventions

## Versioning

- Public HTTP APIs are rooted at `/api/v1`.
- Health and platform endpoints remain unversioned: `/health`, `/ready`, `/metrics` (optional dedicated bind), `/version`, `/openapi.yaml`, `/docs`, `/swagger`.
- Breaking response or request contract changes require a new major API prefix.

## Response Envelope

Success responses:

```json
{
  "success": true,
  "data": {},
  "meta": {},
  "trace_id": "request-id"
}
```

Error responses:

```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "The request contains invalid fields.",
    "fields": [{ "field": "email", "message": "must be a valid email" }]
  },
  "trace_id": "request-id"
}
```

## Pagination

Use page pagination for v1 list endpoints:

- `page`: 1-based page number, default `1`.
- `limit`: page size, default `20`, max `100`.
- Response metadata includes `page`, `limit`, `total`, and `has_next`.

## Dates And Money

- Dates use RFC3339 UTC timestamps.
- Money is represented as integer minor units plus ISO currency code, for example `amount_cents: 1299`, `currency: "USD"`.

## Idempotency

Mutating checkout, payment, and order endpoints require `Idempotency-Key`.
Keys should be unique per user action and safe to retry after network failures.

The API server stores completed responses in **Redis** keyed by scope (authenticated
`X-User-ID` when present, otherwise client IP) plus the idempotency key. Retries and
concurrent duplicate requests receive the same HTTP status, headers, and body as the
first successful response. Redis must be available for these routes.

## Local configuration

At startup the server loads dotenv files depending on the working directory: from the
`backend/` module directory it loads `.env` then `internal/.env` (later overrides); from
the monorepo root it loads `backend/.env` then `backend/internal/.env`. Viper still reads
`BITIK_*` from the process environment (highest precedence).

For repeatable local vs production runs without editing `.env` by hand, use
`make run-local` (copies `.env.local` to `.env`) or `make run-production` (copies `.env.prod`
to `.env`) from the `backend/` directory. Do not commit secret env files.

## V1 Business Defaults

- Countries and cities: configurable; local development defaults are intentionally open.
- Currencies: schema defaults to `USD`, with future support for country-specific currencies.
- Languages: profile default is `en`.
- Taxes: v1 stores `tax_cents`; calculation rules are country-specific and module-owned.
- Seller verification: sellers begin as `pending` and require admin approval before active selling.
- Shipping: v1 supports provider-based shipments with status tracking.
- Payments: v1 supports pay on delivery and manual Wave confirmation; automated providers can be added through payment/webhook modules.

## Status Values

Order status:

- `pending_payment`
- `paid`
- `processing`
- `shipped`
- `delivered`
- `completed`
- `cancelled`
- `refunded`
- `disputed`

Payment status:

- `pending`
- `authorized`
- `paid`
- `failed`
- `refunded`
- `cancelled`

Shipment status:

- `pending`
- `packed`
- `shipped`
- `in_transit`
- `delivered`
- `failed`
- `returned`

Refund and return flows should be represented by order/payment status plus future module-specific audit records.
