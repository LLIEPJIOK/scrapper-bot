package github

import "time"

type Data struct {
	UpdatedAt time.Time `json:"updated_at"`
	Timestamp time.Time `json:"timestamp"`
}
