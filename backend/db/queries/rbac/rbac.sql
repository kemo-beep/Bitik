-- name: ListRoleNamesForUser :many
SELECT r.name
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
ORDER BY r.name ASC;

-- name: AdminListRoles :many
SELECT id, name, description, created_at
FROM roles
ORDER BY name ASC;

-- name: AdminCreateRole :one
INSERT INTO roles (name, description)
VALUES (sqlc.arg('name'), sqlc.narg('description'))
RETURNING id, name, description, created_at;

-- name: AdminUpdateRole :one
UPDATE roles
SET name = COALESCE(sqlc.narg('name')::text, name),
    description = COALESCE(sqlc.narg('description')::text, description)
WHERE id = sqlc.arg('id')
RETURNING id, name, description, created_at;

-- name: AdminDeleteRole :exec
DELETE FROM roles
WHERE id = $1;

-- name: AdminListPermissions :many
SELECT id, key, description, created_at
FROM permissions
ORDER BY key ASC;

-- name: AdminCreatePermission :one
INSERT INTO permissions (key, description)
VALUES (sqlc.arg('key'), sqlc.narg('description'))
RETURNING id, key, description, created_at;

-- name: AdminUpdatePermission :one
UPDATE permissions
SET key = COALESCE(sqlc.narg('key')::text, key),
    description = COALESCE(sqlc.narg('description')::text, description)
WHERE id = sqlc.arg('id')
RETURNING id, key, description, created_at;

-- name: AdminDeletePermission :exec
DELETE FROM permissions
WHERE id = $1;

-- name: AdminListPermissionsForRole :many
SELECT p.id, p.key, p.description, p.created_at
FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
WHERE rp.role_id = $1
ORDER BY p.key ASC;

-- name: AdminAddPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES (sqlc.arg('role_id'), sqlc.arg('permission_id'))
ON CONFLICT DO NOTHING;

-- name: AdminRemovePermissionFromRole :exec
DELETE FROM role_permissions
WHERE role_id = sqlc.arg('role_id')
  AND permission_id = sqlc.arg('permission_id');

-- name: AdminClearPermissionsForRole :exec
DELETE FROM role_permissions
WHERE role_id = $1;

-- name: ListCasbinRules :many
SELECT id, ptype, v0, v1, v2, v3, v4, v5
FROM casbin_rule
ORDER BY id ASC;

-- name: AssignUserRoleByRoleName :exec
INSERT INTO user_roles (user_id, role_id)
SELECT sqlc.arg(user_id), r.id
FROM roles r
WHERE r.name = sqlc.arg(role_name)
ON CONFLICT DO NOTHING;

-- name: RemoveUserRoleByRoleName :exec
DELETE FROM user_roles ur
USING roles r
WHERE ur.role_id = r.id
  AND ur.user_id = sqlc.arg(user_id)
  AND r.name = sqlc.arg(role_name);

-- name: ClearUserRoles :exec
DELETE FROM user_roles
WHERE user_id = $1;
