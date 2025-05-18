-- name: GetUserByEmail :one
SELECT id, created_at, email, password_hash, activated
FROM users
WHERE email = $1;


-- name: CreateUser :one
INSERT INTO users (email, password_hash, activated)
VALUES ($1, $2, $3 )
RETURNING id, created_at;


-- name: UpdateUser :one
UPDATE users
SET email = $1, password_hash = $2, activated = $3
WHERE id = $4
RETURNING id, created_at;


-- name: GetHashTokenForUser :one
SELECT users.id, users.created_at,users.email, users.password_hash,users.activated
FROM users
INNER JOIN token
ON users.id = token.user_id
WHERE token.hash = $1
AND token.scope = $2
AND token.expiry > $3;