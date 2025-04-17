package domain

import "time"

type Update struct {
	ID        int64     `json:"id"         db:"id"`
	ChatID    int64     `json:"chat_id"    db:"chat_id"`
	URL       string    `json:"url"        db:"url"`
	Message   string    `json:"message"    db:"message"`
	Tags      []string  `json:"tags"       db:"tags"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
