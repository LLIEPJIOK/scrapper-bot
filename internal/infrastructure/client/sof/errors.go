package sof

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

type ErrQuestionNotFound struct {
	ID string
}

func NewErrQuestionNotFound(id string) error {
	return ErrQuestionNotFound{
		ID: id,
	}
}

func (e ErrQuestionNotFound) Error() string {
	return fmt.Sprintf("question with id=%q not found", e.ID)
}
