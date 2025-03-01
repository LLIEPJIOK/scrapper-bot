package processor

import (
	"context"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

type Commander struct {
	fsm.BaseTransition
}

func NewCommander() *Commander {
	return &Commander{
		BaseTransition: fsm.BaseTransition{
			Auto: true,
		},
	}
}

func (c *Commander) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	switch state.Message {
	case "/start":
		return &fsm.Result[*State]{
			NextState: start,
			Result:    state,
		}

	default:
		slog.Warn("unknown command", slog.Any("command", state.Message))

		// TODO: handle unknown command
		return &fsm.Result[*State]{
			Result: state,
		}
	}
}
