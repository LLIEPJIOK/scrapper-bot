package bot

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
