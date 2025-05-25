package processor

type ErrFailedState struct {
}

func NewErrFailedState() error {
	return ErrFailedState{}
}

func (e ErrFailedState) Error() string {
	return "failed state"
}
