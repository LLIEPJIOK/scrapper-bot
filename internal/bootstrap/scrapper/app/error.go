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

type ErrUnknownTransport struct {
	Transport string
}

func NewErrUnknownTransport(transport string) error {
	return ErrUnknownTransport{
		Transport: transport,
	}
}

func (e ErrUnknownTransport) Error() string {
	return fmt.Sprintf("unknown transport: %s", e.Transport)
}
