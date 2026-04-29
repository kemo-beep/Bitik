# Phase 11 PII Handling and Deletion Workflow

## PII Classification

- **Direct identifiers:** email, phone, address fields, device identifiers.
- **Sensitive account metadata:** session/user-agent/IP associations.
- **Financially relevant references:** payment identifiers and settlement records.

## Handling Policy

- Minimize PII in logs; avoid plaintext secrets/tokens.
- Store auth tokens hashed (refresh and reset/verify token hash patterns).
- Restrict privileged access through RBAC and admin APIs.

## Deletion Workflow (User-Initiated)

1. User calls account deletion endpoint (`DELETE /api/v1/users/me`).
2. Account is soft-deleted.
3. Refresh tokens and active sessions are revoked.
4. Follow-up operational process can perform irreversible anonymization where required by policy/regulation.

## Operational Extension (Recommended)

- Add scheduled anonymization worker for long-deleted accounts:
  - clear direct identifiers
  - retain non-PII transactional aggregates required for accounting/compliance.
- Maintain legal-hold flag support for disputes/fraud investigations.
