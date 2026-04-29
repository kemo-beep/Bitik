# Web monitoring and observability

## Errors and performance (Sentry)

The app uses `@sentry/nextjs` with `sentry.*.config.ts` in `bitik-web/`. In production:

- Ensure `NEXT_PUBLIC_SENTRY_DSN` is present in the **built** image (build-time for client; server config may use env at runtime per your Sentry setup).
- Use **Sentry Releases** tied to Git SHA or image tag so regressions map to deploys.
- Watch **Issues** (unhandled errors, API failures surfaced to the client) and **Performance** (transaction duration, slow pages).

## Web Vitals

[`bitik-web/components/observability/web-vitals.tsx`](../bitik-web/components/observability/web-vitals.tsx) forwards Core Web Vitals into Sentry breadcrumbs/context. In Sentry, filter by release and URL to spot LCP/CLS regressions after deploys.

## Uptime

The repo does not host Grafana. Recommended:

- Dockploy or edge **HTTP health checks** against `/` (and a future `/api/health` if added).
- External uptime (e.g. synthetic checks) for login and checkout entry points.

## Correlating with the API

- Tag browser sessions or Sentry events with **release** and **environment** (preview / staging / production).
- On the backend, use existing metrics and logs; correlate by **request id**, **user id**, and **timestamp** when investigating checkout or payment failures.

## Product analytics

Client events are documented in [bitik-web/docs/analytics-event-map.md](../bitik-web/docs/analytics-event-map.md). For conversion funnels, export or warehouse those events (or your analytics provider) and chart:

- View product → add to cart → begin checkout → place order
- Seller and admin flows as needed

## What to chart (summary)

| Area | Where |
|------|--------|
| JS errors, release health | Sentry Issues / Releases |
| LCP, INP, CLS trends | Sentry Performance + Web Vitals context |
| HTTP availability | Dockploy / LB + external synthetics |
| API latency and 5xx | Backend dashboards + Sentry linked traces if configured |
| Business funnel | Analytics pipeline from documented client events |
