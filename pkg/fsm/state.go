package fsm

import "context"

type State string

func (s State) String() string {
	return string(s)
}

type StateHandler[TData any] interface {
	Handle(ctx context.Context, data TData) *Result[TData]
	AutoTransition() bool
}

type Result[TData any] struct {
	NextState State
	Result    TData
	Error     error
}
