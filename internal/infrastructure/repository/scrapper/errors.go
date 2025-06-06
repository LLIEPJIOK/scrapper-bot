package scrapper

import "fmt"

type ErrUnknownDBType struct {
	Type string
}

func NewErrUnknownDBType(tpe string) error {
	return ErrUnknownDBType{
		Type: tpe,
	}
}

func (e ErrUnknownDBType) Error() string {
	return fmt.Sprintf("unknown db type: %s", e.Type)
}

type ErrUnregister struct {
	ChatID int64
}

func NewErrUnregister(chatID int64) error {
	return ErrUnregister{
		ChatID: chatID,
	}
}

func (e ErrUnregister) Error() string {
	return fmt.Sprintf("Chat #%d is not registered", e.ChatID)
}

type ErrLinkNotFound struct {
	URL string
}

func NewErrLinkNotFound(url string) error {
	return ErrLinkNotFound{
		URL: url,
	}
}

func (e ErrLinkNotFound) Error() string {
	return fmt.Sprintf("link with url=%q not found", e.URL)
}
