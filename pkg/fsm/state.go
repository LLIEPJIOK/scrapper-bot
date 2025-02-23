package fsm

import "context"

type State string

func (s State) String() string {
	return string(s)
}

type StateHandler[TData any, TResult any] interface {
	Handle(ctx context.Context, data TData, prev *Result[TResult]) *Result[TResult]
	AutoTransition() bool
}

type Result[TResult any] struct {
	NextState *State
	Result    TResult
	Error     error
}
