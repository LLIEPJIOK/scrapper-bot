package fsm

type Builder[TData any, TResult any] struct {
	fsm *FSM[TData, TResult]
}

func NewBuilder[TData any, TResult any]() *Builder[TData, TResult] {
	return &Builder[TData, TResult]{
		fsm: New[TData, TResult](),
	}
}

func (b *Builder[TData, TResult]) AddState(
	state State,
	handler StateHandler[TData, TResult],
) *Builder[TData, TResult] {
	b.fsm.RegisterHandler(state, handler)
	return b
}

func (b *Builder[TData, TResult]) AddTransition(
	from, to State,
) *Builder[TData, TResult] {
	b.fsm.AddTransition(from, to)
	return b
}

func (b *Builder[TData, TResult]) Build() *FSM[TData, TResult] {
	return b.fsm
}
