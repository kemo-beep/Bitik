# Environment Variables Reference

Canonical variable examples live in `backend/.env.example`.

## Critical groups

- `BITIK_DATABASE_*`: primary DB URL and timeout controls.
- `BITIK_REDIS_*`: rate limit, lockout, idempotency support.
- `BITIK_RABBITMQ_*`: async job transport for workers.
- `BITIK_STORAGE_*`: object storage for media.
- `BITIK_SEARCH_*`: OpenSearch endpoint and index behavior.
- `BITIK_AUTH_*`: JWT, lockout, OTP, cookie settings.
- `BITIK_SECURITY_*`: body/upload size limits.
- `BITIK_PAYMENTS_WEBHOOK_SECRET`: webhook verification key.

Do not commit real secrets; use deployment secret stores and CI secrets.
