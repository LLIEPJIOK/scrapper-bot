package domain

type Link struct {
	ID      int64    `json:"id"`
	ChatID  int64    `json:"chat_id"`
	URL     string   `json:"url"`
	Tags    []string `json:"tags"`
	Filters []string `json:"filters"`
}
