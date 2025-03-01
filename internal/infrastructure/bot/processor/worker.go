package processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

func (p *Processor) worker(ctx context.Context, workCh chan State) {
	for state := range workCh {
		res, err := p.ProcessRequest(ctx, &state)
		if err != nil {
			slog.Error("failed to process request",
				slog.Any("current_state", state.FSMState),
				slog.Any("chat_	id", state.ChatID),
				slog.Any("error", err))
		}

		if res != nil {
			res.Result.FSMState = res.NextState
			// TODO: save
		}
	}
}

func (p *Processor) ProcessRequest(
	ctx context.Context,
	state *State,
) (*fsm.Result[*State], error) {
	result := p.fsm.ProcessState(ctx, state.FSMState, state)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to process state: %w", result.Error)
	}

	return result, nil
}
