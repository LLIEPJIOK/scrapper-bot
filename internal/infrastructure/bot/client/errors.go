package client

type ErrResponse struct {
	Message string
}

func NewErrResponse(msg string) error {
	return ErrResponse{
		Message: msg,
	}
}

func (e ErrResponse) Error() string {
	return e.Message
}
