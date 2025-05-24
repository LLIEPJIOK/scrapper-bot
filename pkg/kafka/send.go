package kafka

func Send(ch chan *Input, in *Input) error {
	in.ResChan = make(chan error)
	defer close(in.ResChan)

	ch <- in

	return <-in.ResChan
}
