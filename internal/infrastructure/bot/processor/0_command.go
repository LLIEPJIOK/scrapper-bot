package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

type Commander struct {
}

func NewCommander() *Commander {
	return &Commander{}
}

func (h *Commander) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	switch state.Message {
	case "/start":
		return &fsm.Result[*State]{
			IsAutoTransition: true,
			NextState:        start,
			Result:           state,
		}

	case "/track":
		return &fsm.Result[*State]{
			IsAutoTransition: true,
			NextState:        track,
			Result:           state,
		}

	default:
		state.ShowError = "неопознанная команда"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
		}
	}
}
