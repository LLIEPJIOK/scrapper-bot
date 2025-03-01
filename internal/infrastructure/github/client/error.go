package client

import "fmt"

type ErrInvalidLink struct {
	Link    string
	Message string
}

func NewErrInvalidLink(link, msg string) error {
	return ErrInvalidLink{
		Link:    link,
		Message: msg,
	}
}

func (e ErrInvalidLink) Error() string {
	return fmt.Sprintf("invalid link %q: %s", e.Link, e.Message)
}
