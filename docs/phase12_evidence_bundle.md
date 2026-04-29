# Phase 12 Evidence Bundle

## CI/CD Workflow References

- Backend CI gates: `.github/workflows/backend.yml`
- Web CI + API generation + Playwright: `.github/workflows/web-ci.yml`
- Staging CD: `.github/workflows/backend-cd-staging.yml`
- Production CD + rollback: `.github/workflows/backend-cd-production.yml`

## Test Suites Added

- Unit tests:
  - `backend/internal/promotionsvc/service_test.go`
  - `backend/internal/notificationsvc/service_test.go`
  - `backend/internal/searchsvc/handlers_test.go`
  - `backend/internal/adminsvc/phase9_handlers_test.go`
- Integration tests:
  - `backend/internal/searchsvc/integration_test.go`
  - `backend/internal/mediasvc/integration_test.go`
- API contract checks:
  - `backend/internal/openapi/contract_test.go`
- Playwright critical flows:
  - `bitik-web/tests/phase12-critical-flows.spec.ts`
- Load tests:
  - `backend/loadtests/*.js`

## Coverage and Artifacts

- Coverage profile path in CI artifact: `backend/coverage.out`
- Artifact name: `backend-coverage`

## OpenAPI Contract Output

- Contract test package: `./internal/openapi`
- CI command: `go test ./internal/openapi -run TestOpenAPIMVPContract`

## Runbook Links

- Deployment: `docs/deployment_runbook_phase12.md`
- Payment approvals: `docs/payment_approval_runbook_phase12.md`
- Incident response: `docs/incident_response_runbook_phase12.md`
- Launch checklist: `docs/launch_checklist_phase12.md`
