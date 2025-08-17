-- name: CreateChirp :one
INSERT INTO chirps(created_at, updated_at, user_id)
VALUES (
	$1,
	$2,
	$3
)
RETURNING *;
