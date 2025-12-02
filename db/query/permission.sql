-- name: CreatePermission :one
INSERT INTO permissions (
  name,
  description
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions
WHERE id = $1 LIMIT 1;

-- name: ListPermissions :many
SELECT * FROM permissions
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: CountPermissions :one
SELECT count(*) FROM permissions;

-- name: UpdatePermission :one
UPDATE permissions
SET 
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE id = $1;
