-- +goose Up
-- +goose StatementBegin
CREATE TABLE
	chats (
		id BIGINT PRIMARY KEY,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
		deleted_at TIMESTAMPTZ DEFAULT NULL
	);

CREATE TABLE
	links (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
	);

CREATE TABLE
	tags (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		name TEXT NOT NULL UNIQUE
	);

CREATE TABLE
	filters (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		value TEXT NOT NULL UNIQUE
	);

CREATE TABLE
	links_tags (
		link_id BIGINT NOT NULL REFERENCES links (id),
		tag_id BIGINT NOT NULL REFERENCES tags (id),
		chat_id BIGINT NOT NULL REFERENCES chats (id),
		PRIMARY KEY (link_id, tag_id)
	);

CREATE TABLE
	links_filters (
		link_id BIGINT NOT NULL REFERENCES links (id),
		filter_id BIGINT NOT NULL REFERENCES filters (id),
		chat_id BIGINT NOT NULL REFERENCES chats (id),
		PRIMARY KEY (link_id, filter_id)
	);

CREATE TABLE
	links_chats (
		link_id BIGINT NOT NULL REFERENCES links (id),
		chat_id BIGINT NOT NULL REFERENCES chats (id),
		PRIMARY KEY (link_id, chat_id)
	);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE links_chats;

DROP TABLE links_filters;

DROP TABLE links_tags;

DROP TABLE filters;

DROP TABLE tags;

DROP TABLE links;

-- +goose StatementEnd