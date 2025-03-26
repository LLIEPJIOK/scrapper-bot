package fsm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
)

type FSM[TData any] struct {
	mu          sync.RWMutex
	handlers    map[string]StateHandler[TData]
	transitions map[string][]string
}

func New[TData any]() *FSM[TData] {
	return &FSM[TData]{
		handlers:    make(map[string]StateHandler[TData]),
		transitions: make(map[string][]string),
	}
}

func (w *FSM[TData]) RegisterHandler(
	state State,
	handler StateHandler[TData],
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.handlers[state.String()] = handler
}

func (w *FSM[TData]) AddTransition(from, to State) {
	w.mu.Lock()
	defer w.mu.Unlock()

	fromStr := from.String()
	toStr := to.String()
	w.transitions[fromStr] = append(w.transitions[fromStr], toStr)
}

func (w *FSM[TData]) CanTransition(from, to State) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	fromStr := from.String()
	toStr := to.String()
	validTransitions := w.transitions[fromStr]

	return slices.Contains(validTransitions, toStr)
}

func (w *FSM[TData]) ProcessState(
	ctx context.Context,
	state State,
	data TData,
) (result *Result[TData]) {
	defer func() {
		if err := recover(); err != nil {
			result = &Result[TData]{
				Error: errors.Join(result.Error, fmt.Errorf("panic: %v", err)),
			}
		}
	}()

	w.mu.RLock()
	handler, exists := w.handlers[state.String()]
	w.mu.RUnlock()

	if !exists {
		result = &Result[TData]{
			Error: fmt.Errorf("no handler registered for state %s", state),
		}

		return result
	}

	result = handler.Handle(ctx, data)
	if result.Error != nil {
		slog.Error("handle error", slog.Any("error", result.Error), slog.Any("fsm_state", state))
	}

	if result.IsAutoTransition {
		if !w.CanTransition(state, result.NextState) {
			result = &Result[TData]{
				Error: fmt.Errorf(
					"invalid state transition from %s to %s",
					state,
					result.NextState,
				),
			}

			return result
		}

		return w.ProcessState(ctx, result.NextState, result.Result)
	}

	return result
}
