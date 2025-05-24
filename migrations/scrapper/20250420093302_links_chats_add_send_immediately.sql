-- +goose Up
-- +goose StatementBegin
ALTER TABLE links_chats
ADD COLUMN send_immediately BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links_chats
DROP COLUMN send_immediately;
-- +goose StatementEnd