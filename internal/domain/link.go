package domain

import "time"

type Link struct {
	ID      int64    `json:"id"`
	ChatID  int64    `json:"chat_id"`
	URL     string   `json:"url"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}

type CheckLink struct {
	ID        int64
	URL       string
	Filters   []string
	Tags      []string
	Chats     []int64
	CheckedAt time.Time
}
