package domain

import "time"

type Link struct {
	ID              int64      `json:"id"               db:"id"`
	ChatID          int64      `json:"chat_id"          db:"chat_id"`
	URL             string     `json:"url"              db:"url"`
	Tags            []string   `json:"tags"             db:"tags"`
	Filters         []string   `json:"filters"          db:"filters"`
	SendImmediately Null[bool] `json:"send_immediately" db:"send_immediately"`
}

type CheckLink struct {
	ID        int64      `json:"id"         db:"id"`
	URL       string     `json:"url"        db:"url"`
	Chats     []LinkChat `json:"chats"      db:"chats"`
	CheckedAt time.Time  `json:"checked_at" db:"checked_at"`
}

type LinkChat struct {
	ChatID          int64    `json:"chat_id"          db:"chat_id"`
	SendImmediately bool     `json:"send_immediately" db:"send_immediately"`
	Filters         []string `json:"filters"          db:"filters"`
	Tags            []string `json:"tags"             db:"tags"`
}
