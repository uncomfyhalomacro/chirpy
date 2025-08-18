-- name: AddRefreshToken :one
INSERT INTO refresh_tokens (
	token,
	created_at,
	updated_at,
	expires_at,
	user_id
) VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *;

-- name: GetExpiry :one
SELECT expires_at, revoked_at FROM refresh_tokens
WHERE token=$1 AND user_id=$2 LIMIT 1;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
WHERE id=(
	SELECT refresh_tokens.user_id FROM refresh_tokens
	WHERE token=$1 LIMIT 1
) LIMIT 1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at=$2, updated_at=$2
WHERE token=$1;

