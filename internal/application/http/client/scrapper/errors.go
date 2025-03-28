package scrapper

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

type ErrUserResponse struct {
	Message string
}

func NewErrUserResponse(msg string) error {
	return ErrUserResponse{
		Message: msg,
	}
}

func (e ErrUserResponse) Error() string {
	return e.Message
}
