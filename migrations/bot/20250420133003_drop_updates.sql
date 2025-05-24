-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_updates_created_at;

DROP TABLE IF EXISTS updates;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE updates (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	chat_id BIGINT NOT NULL,
	url TEXT NOT NULL,
	message TEXT NOT NULL,
	tags TEXT[] DEFAULT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);

CREATE INDEX idx_updates_created_at ON updates(created_at);
-- +goose StatementEnd
