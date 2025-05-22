-- name: CreateTransaction :exec
INSERT INTO transaction (type, category_id, amount, date, description) VALUES ($1, $2, $3, $4, $5);

-- name: GetTransaction :one
SELECT transaction.*, category.name as category_name 
FROM transaction JOIN category ON transaction.category_id = category.id WHERE transaction.id = $1;

-- name: GetTransactions :many
SELECT transaction.*, category.name as category_name 
FROM transaction JOIN category ON transaction.category_id = category.id;

-- name: UpdateTransaction :exec
UPDATE transaction SET type = $2, category_id = $3, amount = $4, date = $5, description = $6 WHERE id = $1;

-- name: DeleteTransaction :exec
DELETE FROM transaction WHERE id = $1;

-- name: GetTransactionsByType :many
SELECT transaction.*, category.name as category_name 
FROM transaction JOIN category ON transaction.category_id = category.id
WHERE transaction.type = $1 ORDER BY transaction.created_at DESC;
