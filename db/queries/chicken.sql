-- name: InsertChicken :exec
INSERT INTO chicken (type, quantity) VALUES ($1, $2);

-- name: GetChicken :one
SELECT * FROM chicken WHERE type = $1;

-- name: GetChickens :many
SELECT * FROM chicken;

-- name: UpdateChickenById :exec
UPDATE chicken SET quantity = $2 WHERE id = $1;

-- name: DeleteChicken :exec
DELETE FROM chicken WHERE type = $1;
