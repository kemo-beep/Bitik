# Payment Approval Runbook (Wave Manual + POD)

## Wave Manual

- Admin queue page: `/admin/payments/wave`.
- Review pending payment evidence.
- Approve or reject using admin actions:
  - `POST /api/v1/admin/payments/{payment_id}/wave/approve`
  - `POST /api/v1/admin/payments/{payment_id}/wave/reject`
- Every action should produce admin activity logs and payment status transitions.

## POD Capture

- Trigger capture only when shipment and order state satisfy POD eligibility checks.
- API path: `POST /api/v1/admin/payments/{payment_id}/pod/capture`.
- Verify post-capture jobs enqueue (invoice + wallet settlement).

## Verification Checklist

- Payment row status updated.
- Associated order state consistent.
- Admin activity log exists.
- Follow-up job entries/worker logs present.
