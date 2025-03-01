package fsm

type Builder[TData any] struct {
	fsm *FSM[TData]
}

func NewBuilder[TData any]() *Builder[TData] {
	return &Builder[TData]{
		fsm: New[TData](),
	}
}

func (b *Builder[TData]) AddState(
	state State,
	handler StateHandler[TData],
) *Builder[TData] {
	b.fsm.RegisterHandler(state, handler)
	return b
}

func (b *Builder[TData]) AddTransition(
	from, to State,
) *Builder[TData] {
	b.fsm.AddTransition(from, to)
	return b
}

func (b *Builder[TData]) Build() *FSM[TData] {
	return b.fsm
}
