-- name: GetRolesForUser :many
SELECT roles.* FROM roles
JOIN user_roles ON roles.id = user_roles.role_id
WHERE user_roles.user_id = $1;

-- name: AddRoleForUser :one
INSERT INTO user_roles (
  user_id,
  role_id
) VALUES (
  $1, $2
) RETURNING *;

-- name: RemoveRoleForUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;

-- name: ListUserRoles :many
SELECT * FROM user_roles
ORDER BY user_id, role_id
LIMIT $1
OFFSET $2;

-- name: CountUserRoles :one
SELECT count(*) FROM user_roles;
