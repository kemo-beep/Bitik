# RBAC matrix (Casbin + SQL roles)

## SQL layer

- `roles`, `permissions`, `role_permissions`, `user_roles` model coarse entitlements and UI/feature flags.
- Seed roles: `admin`, `seller`, `buyer` (see `db/seeds/001_development_seed.sql`).

## HTTP layer (Casbin)

Policies live in `casbin_rule` (`ptype = p`). Subject is the **role name**; JWTs carry the same names in the `roles` claim.

| Role   | Path prefix (object)   | Action |
|--------|-------------------------|--------|
| admin  | `/api/v1/admin`         | `*`    |
| admin  | `/api/v1/admin/*`       | `*`    |
| seller | `/api/v1/seller`       | `*`    |
| seller | `/api/v1/seller/*`     | `*`    |
| buyer  | `/api/v1/buyer`        | `*`    |
| buyer  | `/api/v1/buyer/*`      | `*`    |
| buyer  | `/api/v1/users`        | `*`    |
| buyer  | `/api/v1/users/*`      | `*`    |
| seller | `/api/v1/users`        | `*`    |
| seller | `/api/v1/users/*`      | `*`    |
| admin  | `/api/v1/users`        | `*`    |
| admin  | `/api/v1/users/*`      | `*`    |

Matcher: `keyMatch2` on path; action `*` matches any HTTP method.

## Extending

1. Insert new rows into `casbin_rule` (or migrate them).
2. Restart API (policies load at startup), or add a future admin endpoint to reload the enforcer.
