package client

type ErrResponse struct {
	message string
}

func NewErrResponse(msg string) error {
	return ErrResponse{
		message: msg,
	}
}

func (e ErrResponse) Error() string {
	return e.message
}
