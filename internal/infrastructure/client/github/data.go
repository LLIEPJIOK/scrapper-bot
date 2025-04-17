package github

import "time"

type Data struct {
	Title     string      `json:"title"`
	Body      string      `json:"body"`
	Number    int         `json:"number"`
	URL       string      `json:"html_url"`
	User      User        `json:"user"`
	CreatedAt time.Time   `json:"created_at"`
	PR        PullRequest `json:"pull_request"`
}

type User struct {
	Login string `json:"login"`
}

type PullRequest struct {
	URL string `json:"url"`
}
