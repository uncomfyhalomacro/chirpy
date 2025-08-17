-- name: CreateUser :one
INSERT INTO users(created_at, updated_at, email, hashed_password)
VALUES (
	$1,
	$2,
	$3,
	$4
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE email=$1 LIMIT 1;
