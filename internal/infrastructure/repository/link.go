package repository

import "time"

type Link struct {
	ID        int64
	URL       string
	Filters   []string
	Tags      []string
	Chats     []int64
	CheckedAt time.Time
}
