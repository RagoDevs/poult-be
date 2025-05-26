-- name: GetUserByEmail :one
SELECT id, created_at, name, email, password_hash, activated
FROM users
WHERE email = $1;


-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, activated)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at;


-- name: UpdateUser :one
UPDATE users
SET name = $1, email = $2, password_hash = $3, activated = $4
WHERE id = $5
RETURNING id, created_at;


-- name: GetHashTokenForUser :one
SELECT users.id, users.created_at,users.email, users.name, users.password_hash,users.activated
FROM users
INNER JOIN token
ON users.id = token.user_id
WHERE token.hash = $1
AND token.scope = $2
AND token.expiry > $3;