package config

import "fmt"

type ErrConfig struct {
	message string
}

func NewErrConfig(msg string) error {
	return ErrConfig{
		message: msg,
	}
}

func (e ErrConfig) Error() string {
	return fmt.Sprintf("config error: %s", e.message)
}
