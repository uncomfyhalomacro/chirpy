-- +goose Up
CREATE TABLE chirps (
	id	UUID PRIMARY KEY DEFAULT gen_random_uuid (),
	created_at	TIMESTAMP	NOT NULL,
	updated_at	TIMESTAMP	NOT NULL,
	user_id		UUID 		NOT NULL,
	CONSTRAINT FK_user_id
	FOREIGN KEY(user_id)	REFERENCES users(id)
	ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;
