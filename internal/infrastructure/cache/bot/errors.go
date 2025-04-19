package bot

type ErrNoData struct {
}

func NewErrNoData() error {
	return ErrNoData{}
}

func (e ErrNoData) Error() string {
	return "no data"
}
