# Incident Response Runbook (Phase 12)

## Detection

- Alerting sources: API health, worker lag, queue depth, DB connectivity, error rates.
- Initial triage severity:
  - Sev1: checkout/order/payment unavailable
  - Sev2: degraded search/notifications/admin
  - Sev3: non-critical feature regression

## Immediate Actions

1. Confirm impact using `/health`, `/ready`, and synthetic checkout smoke.
2. Check latest deploy and migrations.
3. Review logs/metrics:
   - API + worker error logs
   - queue retries/DLQ metrics
   - DB saturation and lock waits
4. If deploy-related, rollback to last-known-good tag.

## Stabilization

- Disable risky features via `feature_flags` if needed.
- Pause noisy worker consumers if backlog is causing cascading failures.
- Apply hotfix behind targeted validation and smoke checks.

## Closure

- Record incident timeline and root cause.
- Add preventive tests and/or monitors.
- Publish post-incident action items in engineering backlog.
