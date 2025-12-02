-- name: CreateMedicine :one
INSERT INTO medicines (
  name,
  unit,
  price,
  stock,
  description
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetMedicine :one
SELECT * FROM medicines
WHERE id = $1 LIMIT 1;

-- name: ListMedicines :many
SELECT * FROM medicines
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: CountMedicines :one
SELECT count(*) FROM medicines;

-- name: UpdateMedicine :one
UPDATE medicines
SET
  name = COALESCE(sqlc.narg(name), name),
  unit = COALESCE(sqlc.narg(unit), unit),
  price = COALESCE(sqlc.narg(price), price),
  stock = COALESCE(sqlc.narg(stock), stock),
  description = COALESCE(sqlc.narg(description), description),
  updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMedicine :exec
DELETE FROM medicines
WHERE id = $1;
