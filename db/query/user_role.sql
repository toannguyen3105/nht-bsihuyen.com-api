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
