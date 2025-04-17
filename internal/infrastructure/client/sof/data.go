package sof

type Data struct {
	Items []Item `json:"items"`
}

type Item struct {
	Owner User `json:"owner"`
}

type QuestionData struct {
	Items []Question `json:"items"`
}

type Question struct {
	Title string `json:"title"`
}

type AnswerData struct {
	Items []Answer `json:"items"`
}

type Answer struct {
	ID        int64  `json:"answer_id"`
	CreatedAt int64  `json:"creation_date"`
	Owner     User   `json:"owner"`
	Body      string `json:"body"`
}

type CommentData struct {
	Items []Comment `json:"items"`
}

type Comment struct {
	CreatedAt int64  `json:"creation_date"`
	Owner     User   `json:"owner"`
	Body      string `json:"body"`
}

type User struct {
	DisplayName string `json:"display_name"`
}
