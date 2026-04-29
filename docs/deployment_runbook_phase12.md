# Deployment Runbook (Docker Host)

## Prerequisites

- Docker + Docker Compose installed on target host.
- Environment file with production variables available on host.
- SSH access from GitHub Actions runner (or self-hosted runner) to host.
- Secrets configured for staging/production workflow files.

## Staging Deploy

1. Run `.github/workflows/backend-cd-staging.yml`.
2. Workflow performs migration gate (`go run ./cmd/migrate up` + status).
3. Workflow runs `backend/scripts/deploy.sh` with `IMAGE_TAG=${GITHUB_SHA}`.
4. Workflow runs `backend/scripts/smoke.sh` against staging URL.

## Production Deploy

1. Trigger `.github/workflows/backend-cd-production.yml`.
2. Workflow performs migration gate.
3. Workflow deploys image using `backend/scripts/deploy.sh`.
4. Workflow runs smoke tests.
5. If smoke fails, workflow runs `backend/scripts/rollback.sh` using `PROD_LAST_KNOWN_GOOD_TAG`.

## Manual Rollback

```bash
cd backend
DEPLOY_HOST=... DEPLOY_USER=... DEPLOY_PATH=... ROLLBACK_IMAGE_TAG=<known-good-tag> bash ./scripts/rollback.sh
```
