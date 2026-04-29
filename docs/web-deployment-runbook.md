# Web deployment runbook (Bitik · Dockploy + GHCR)

This runbook covers building the Next.js app as a Docker image, publishing to GitHub Container Registry (GHCR), and running it on Dockploy.

## Image location and tags

Images are pushed to:

`ghcr.io/<github-owner-lowercase>/bitik-web`

| Workflow | Trigger | Tags |
|----------|---------|------|
| [web-cd-preview.yml](../.github/workflows/web-cd-preview.yml) | PR touching `bitik-web/` | `pr-<number>-<7-char-sha>` |
| [web-cd-staging.yml](../.github/workflows/web-cd-staging.yml) | Push to `staging` or manual | `staging-<sha>`, `staging-latest` |
| [web-cd-production.yml](../.github/workflows/web-cd-production.yml) | Push `main`/`master`, GitHub Release, manual | `prod-<sha>`, `latest`, plus release tag when applicable |

Prefer **immutable tags** (`pr-*`, `staging-*`, `prod-*`, or release tag) for rollbacks. Treat `staging-latest` / `latest` as pointers only.

## Required build-time environment (`NEXT_PUBLIC_*`)

Next.js bakes public env into the client bundle at **build** time. The Docker build runs [`bitik-web/scripts/verify-web-build-env.mjs`](../bitik-web/scripts/verify-web-build-env.mjs), which requires valid URLs for:

- `NEXT_PUBLIC_API_BASE_URL`
- `NEXT_PUBLIC_ASSET_BASE_URL`
- `NEXT_PUBLIC_WS_BASE_URL`
- `NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL`

Optional (defaults documented in [`bitik-web/lib/env.ts`](../bitik-web/lib/env.ts)):

- `NEXT_PUBLIC_SENTRY_DSN`
- `NEXT_PUBLIC_FEATURE_FLAGS` (JSON object string)
- `NEXT_PUBLIC_ANALYTICS_ENABLED`

### GitHub Actions configuration

**Preview (PR):** Optional repository **Variables** `PREVIEW_NEXT_PUBLIC_*` override defaults baked in the workflow. If unset, placeholder `*.preview.localhost` URLs are used so the image builds; override for a real preview API.

**Staging:** GitHub Environment `staging` — set **Secrets**:

- `STAGING_NEXT_PUBLIC_API_BASE_URL`
- `STAGING_NEXT_PUBLIC_ASSET_BASE_URL`
- `STAGING_NEXT_PUBLIC_WS_BASE_URL`
- `STAGING_NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL`
- `STAGING_NEXT_PUBLIC_SENTRY_DSN` (optional)

Optional **Variables**: `STAGING_NEXT_PUBLIC_FEATURE_FLAGS`, `STAGING_NEXT_PUBLIC_ANALYTICS_ENABLED`.

**Production:** GitHub Environment `production` — same names with `PRODUCTION_` prefix (see workflow file).

**GHCR pull from Dockploy:** Create a read-only GitHub PAT with `read:packages`, or use Dockploy’s GitHub integration. The deploy host must `docker login ghcr.io` (or equivalent) before pull.

## Dockploy wiring

1. **Image:** `ghcr.io/<owner>/bitik-web`
2. **Tag:** Pin staging/prod to immutable tags; use `latest` only if you accept drift.
3. **Port:** Container listens on **3000** (`PORT` / `HOSTNAME` set in Dockerfile). For local Playwright against a production build, run **`npm run build`** then **`npm run start:standalone`** from `bitik-web/`: `postbuild` copies `.next/static` and `public` into `.next/standalone`, and the script runs `node server.js` with cwd `.next/standalone` (same layout as the Docker image). Do not run `node .next/standalone/server.js` from the repo root without those copies or chunks will 404.
4. **Health check:** HTTP GET `/` (200). Optionally add a dedicated route later (e.g. `/api/health`) and point health checks there.
5. **Runtime secrets** (Sentry server, etc.) are separate from `NEXT_PUBLIC_*`; configure in Dockploy if the app gains server-only env.

## Rollback

1. In GHCR, find the previous digest or immutable tag (`prod-<old-sha>` or release tag).
2. In Dockploy, redeploy the service using that tag (or digest).
3. Avoid relying solely on mutating `latest` without recording the previous digest.

## Common failures

| Symptom | Likely cause |
|---------|----------------|
| Build fails at `verify-web-build-env` | Missing or invalid URL in build-args |
| Blank API calls in browser | Wrong `NEXT_PUBLIC_API_BASE_URL` for the environment; rebuild image |
| `_next/static` 404 / broken client JS locally | Run `npm run build` so `postbuild` runs, then `npm run start:standalone` (not `node .next/standalone/server.js` from repo root) |
| OAuth redirect mismatch | `NEXT_PUBLIC_OAUTH_REDIRECT_BASE_URL` must match the public web URL |
| PR preview image push denied | Fork PRs: `GITHUB_TOKEN` may not have `packages:write` for fork workflows; build from same-repo branches or use a PAT workflow |
| WebSocket errors | `NEXT_PUBLIC_WS_BASE_URL` must match deployed WS (scheme `ws`/`wss`) |

## Related docs

- [web-launch-checklist.md](./web-launch-checklist.md)
- [web-monitoring.md](./web-monitoring.md)
