package client

import "fmt"

type ErrInvalidLink struct {
	Link string
}

func NewErrInvalidLink(link string) error {
	return ErrInvalidLink{
		Link: link,
	}
}

func (e ErrInvalidLink) Error() string {
	return fmt.Sprintf("invalid stackoverflow link: %q", e.Link)
}
