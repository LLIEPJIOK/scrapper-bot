package client

import "fmt"

type ErrServiceUnavailable struct {
	Inner error
}

func NewErrServiceUnavailable(inner error) error {
	return ErrServiceUnavailable{
		Inner: inner,
	}
}

func (e ErrServiceUnavailable) Error() string {
	return fmt.Sprintf("service unavailable: %v", e.Inner)
}

func (e ErrServiceUnavailable) Unwrap() error {
	return e.Inner
}
