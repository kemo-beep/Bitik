# Web launch checklist

Use before pointing production traffic at the Dockploy-hosted Next app.

## DNS and TLS

- [ ] Apex and `www` (if used) resolve to the load balancer or Dockploy ingress.
- [ ] TLS certificate valid for the public hostname; HTTP→HTTPS redirect enabled.

## Application URLs and CORS

- [ ] `NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL` matches the **public** storefront URL (scheme + host, no trailing slash issues in OAuth provider config).
- [ ] `NEXT_PUBLIC_API_BASE_URL` points to the production API (`…/api/v1` as generated clients expect).
- [ ] Backend CORS allows the production web origin if the API enforces CORS.

## Build-time public config (rebuild if changed)

- [ ] `NEXT_PUBLIC_ASSET_BASE_URL` for CDN or API-served assets.
- [ ] `NEXT_PUBLIC_WS_BASE_URL` uses `wss://` in production.
- [ ] `NEXT_PUBLIC_FEATURE_FLAGS` JSON reviewed for launch (no accidental dev flags).

## Observability

- [ ] `NEXT_PUBLIC_SENTRY_DSN` set for production build; source maps / release tracking configured in Sentry (see [web-monitoring.md](./web-monitoring.md)).
- [ ] `NEXT_PUBLIC_ANALYTICS_ENABLED` set per product decision; events aligned with [bitik-web/docs/analytics-event-map.md](../bitik-web/docs/analytics-event-map.md).

## Smoke after deploy

- [ ] `/` loads and key navigation works.
- [ ] Register or login (per your auth rollout).
- [ ] Browse → product → add to cart → checkout start (no 401/500 on critical APIs).
- [ ] Admin/staff routes gated as expected.

## Operations

- [ ] Runbook reviewed: [web-deployment-runbook.md](./web-deployment-runbook.md).
- [ ] Rollback path tested once (redeploy previous GHCR tag).
