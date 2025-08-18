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

-- name: UpdateUserDetails :one
UPDATE users
SET email=$1, hashed_password=$2
WHERE id=$3
RETURNING *;

-- name: UpgradeUserToChirpyRed :one
UPDATE users
SET is_chirpy_red=true
WHERE id=$1
RETURNING *;
