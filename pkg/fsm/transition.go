package fsm

type BaseTransition struct {
	Auto bool
}

func (t *BaseTransition) AutoTransition() bool {
	return t.Auto
}
