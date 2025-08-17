-- name: CreateChirp :one
INSERT INTO chirps(body, created_at, updated_at, user_id)
VALUES (
	$1,
	$2,
	$3,
	$4
)
RETURNING *;
