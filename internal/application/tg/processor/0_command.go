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

func (h *Commander) Handle(_ context.Context, state *State) *fsm.Result[*State] {
	switch state.Message {
	case "/start":
		return &fsm.Result[*State]{
			NextState:        start,
			IsAutoTransition: true,
			Result:           state,
		}

	case "/help":
		return &fsm.Result[*State]{
			NextState:        help,
			IsAutoTransition: true,
			Result:           state,
		}

	case "/track":
		return &fsm.Result[*State]{
			NextState:        track,
			IsAutoTransition: true,
			Result:           state,
		}

	case "/untrack":
		return &fsm.Result[*State]{
			NextState:        untrack,
			IsAutoTransition: true,
			Result:           state,
		}

	case "/list":
		return &fsm.Result[*State]{
			NextState:        list,
			IsAutoTransition: true,
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
