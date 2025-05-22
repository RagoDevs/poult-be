-- name: InsertCategory :exec
INSERT INTO category (name, description) VALUES ($1, $2);

-- name: GetCategory :one
SELECT * FROM category WHERE id = $1;

-- name: GetCategoryByName :one
SELECT * FROM category WHERE name = $1;

-- name: GetCategories :many
SELECT * FROM category;

-- name: UpdateCategory :exec
UPDATE category SET name = $2, description = $3 WHERE id = $1;

-- name: DeleteCategory :exec
DELETE FROM category WHERE id = $1;
