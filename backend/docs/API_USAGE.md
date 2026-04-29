# API Usage Reference

- Source of truth: `backend/internal/openapi/openapi.yaml`
- Runtime endpoint:
  - `GET /openapi.yaml`
  - `GET /docs`

## Regeneration and Drift Checks

- Web types generation:
  - `cd bitik-web && npm run api:gen`
- Backend contract assertion:
  - `cd backend && go test ./internal/openapi -run TestOpenAPIMVPContract`
