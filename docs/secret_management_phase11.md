# Phase 11 Secret Management Guidance

## Approved Sources

- Doppler, Infisical, or cloud-native secret manager (AWS/GCP/Azure).

## Rules

- Never commit secrets to repository, plan files, or examples.
- Keep `.env.example` non-sensitive and placeholder-only.
- Inject production secrets at runtime through environment variables.

## Rotation

- Rotate JWT secret and webhook secrets on defined cadence.
- Rotate DB and storage credentials with staged rollout and rollback plan.

## Minimal Required Secrets

- `BITIK_AUTH_JWT_SECRET`
- `BITIK_DATABASE_URL`
- `BITIK_STORAGE_ACCESS_KEY_ID`
- `BITIK_STORAGE_SECRET_ACCESS_KEY`
- `BITIK_PAYMENTS_WEBHOOK_SECRET`
