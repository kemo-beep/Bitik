# Phase 11 Backup and Restore Runbook

## PostgreSQL Backup

- **Full backup (daily):**
  - `pg_dump -Fc "$BITIK_DATABASE_URL" > bitik-$(date +%F).dump`
- **Schema-only backup (per release):**
  - `pg_dump -s "$BITIK_DATABASE_URL" > bitik-schema-$(date +%F).sql`
- Store artifacts in encrypted object storage with lifecycle retention.

## PostgreSQL Restore

- Create clean target DB.
- Restore:
  - `pg_restore -d "$TARGET_DB_URL" bitik-YYYY-MM-DD.dump`
- Validate:
  - run migrations status
  - run integrity smoke checks (users/orders/payments counts).

## Object Storage Metadata Backup

- Persist metadata tables in PostgreSQL (`media_files`, related ownership references).
- For binary objects, enable bucket versioning and lifecycle policies at storage layer.
- Keep periodic object inventory export in secure archive.

## Disaster Recovery Drill (Monthly)

- Restore latest backup into staging.
- Run application smoke tests and key API checks.
- Record RTO/RPO and remediation items.

## Secrets + Credentials

- Never store DB credentials or access keys in repo.
- Use secret manager-backed environment variables for backup automation jobs.
