package fsm

import (
	"context"
	"fmt"
	"sync"
)

type FSM[TData any, TResult any] struct {
	mu          sync.RWMutex
	handlers    map[string]StateHandler[TData, TResult]
	transitions map[string][]string
}

func New[TData any, TResult any]() *FSM[TData, TResult] {
	return &FSM[TData, TResult]{
		handlers:    make(map[string]StateHandler[TData, TResult]),
		transitions: make(map[string][]string),
	}
}

func (w *FSM[TData, TResult]) RegisterHandler(
	state State,
	handler StateHandler[TData, TResult],
) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[state.String()] = handler
}

func (w *FSM[TData, TResult]) AddTransition(from, to State) {
	w.mu.Lock()
	defer w.mu.Unlock()

	fromStr := from.String()
	toStr := to.String()

	if _, exists := w.transitions[fromStr]; !exists {
		w.transitions[fromStr] = make([]string, 0)
	}
	w.transitions[fromStr] = append(w.transitions[fromStr], toStr)
}

func (w *FSM[TData, TResult]) CanTransition(from, to *State) bool {
	if from == nil || to == nil {
		return false
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	fromStr := (*from).String()
	toStr := (*to).String()

	validTransitions, exists := w.transitions[fromStr]
	if !exists {
		return false
	}

	for _, state := range validTransitions {
		if state == toStr {
			return true
		}
	}
	return false
}

func (w *FSM[TData, TResult]) ProcessState(
	ctx context.Context,
	curState *State,
	data TData,
	prev *Result[TResult],
) *Result[TResult] {
	w.mu.RLock()
	handler, exists := w.handlers[(*curState).String()]
	w.mu.RUnlock()

	var result *Result[TResult]

	if !exists {
		result = &Result[TResult]{
			Error: fmt.Errorf("no handler registered for state %s", curState),
		}

		return result
	}

	result = handler.Handle(ctx, data, prev)

	if handler.AutoTransition() && result.NextState != nil {
		if !w.CanTransition(curState, result.NextState) {
			result = &Result[TResult]{
				Error: fmt.Errorf(
					"invalid state transition from %s to %s",
					curState,
					result.NextState,
				),
			}

			return result
		}

		return w.ProcessState(ctx, result.NextState, data, result)
	}

	return result
}
