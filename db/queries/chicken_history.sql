-- name: InsertChickenHistory :exec
INSERT INTO chicken_history (chicken_type, quantity_change, reason) VALUES ($1, $2, $3);

-- name: GetChickenHistory :one
SELECT * FROM chicken_history WHERE id = $1;

-- name: GetChickenHistories :many
SELECT * FROM chicken_history;

-- name: UpdateChickenHistory :exec
UPDATE chicken_history SET chicken_type = $2, quantity_change = $3, reason = $4 WHERE id = $1;

-- name: DeleteChickenHistory :exec
DELETE FROM chicken_history WHERE id = $1;