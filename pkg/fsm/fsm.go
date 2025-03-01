package fsm

import (
	"context"
	"errors"
	"fmt"
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
	curState State,
	data TData,
) (result *Result[TData]) {
	defer func() {
		if err := recover(); err != nil {
			result.Error = errors.Join(result.Error, fmt.Errorf("panic: %v", err))
		}
	}()

	w.mu.RLock()
	handler, exists := w.handlers[curState.String()]
	w.mu.RUnlock()

	if !exists {
		result = &Result[TData]{
			Error: fmt.Errorf("no handler registered for state %s", curState),
		}

		return result
	}

	result = handler.Handle(ctx, data)

	if result.IsAutoTransition {
		if !w.CanTransition(curState, result.NextState) {
			result = &Result[TData]{
				Error: fmt.Errorf(
					"invalid state transition from %s to %s",
					curState,
					result.NextState,
				),
			}

			return result
		}

		return w.ProcessState(ctx, result.NextState, result.Result)
	}

	return result
}
