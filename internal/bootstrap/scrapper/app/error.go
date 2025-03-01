package app

import "fmt"

type ErrStopApp struct {
	Message string
}

func NewErrStopApp(msg string) error {
	return ErrStopApp{
		Message: msg,
	}
}

func (e ErrStopApp) Error() string {
	return fmt.Sprintf("failed to stop app: %s", e.Message)
}
