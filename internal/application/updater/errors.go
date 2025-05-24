package updater

type ErrSendUpdate struct {
}

func NewErrSendUpdate() error {
	return ErrSendUpdate{}
}

func (e ErrSendUpdate) Error() string {
	return "failed to send update: all handlers are unavailable"
}
