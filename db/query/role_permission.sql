-- name: CreateRolePermission :one
INSERT INTO role_permissions (
  role_id,
  permission_id
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetRolePermission :one
SELECT * FROM role_permissions
WHERE role_id = $1 AND permission_id = $2
LIMIT 1;

-- name: DeleteRolePermission :exec
DELETE FROM role_permissions
WHERE role_id = $1 AND permission_id = $2;

-- name: ListRolePermissions :many
SELECT * FROM role_permissions
ORDER BY role_id, permission_id
LIMIT $1
OFFSET $2;

-- name: CountRolePermissions :one
SELECT count(*) FROM role_permissions;

-- name: UpdateRolePermission :one
UPDATE role_permissions
SET 
  permission_id = $3,
  updated_at = now()
WHERE role_id = $1 AND permission_id = $2
RETURNING *;
