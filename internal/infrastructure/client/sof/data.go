package sof

type Data struct {
	Items []Item `json:"items"`
}

type Item struct {
	Owner User `json:"owner"`
}

type User struct {
	DisplayName string `json:"display_name"`
}
