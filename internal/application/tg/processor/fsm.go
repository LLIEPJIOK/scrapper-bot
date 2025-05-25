package processor

import (
	"context"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

type CustomFSM struct {
	core    *fsm.FSM[*State]
	metrics Metrics
}

func NewCustomFSM(core *fsm.FSM[*State], metrics Metrics) *CustomFSM {
	return &CustomFSM{
		core:    core,
		metrics: metrics,
	}
}

func (f *CustomFSM) ProcessState(
	ctx context.Context,
	state fsm.State,
	data *State,
) *fsm.Result[*State] {
	start := time.Now()
	res := f.core.ProcessState(ctx, state, data)
	duration := time.Since(start)

	f.metrics.ObserveTGRequestsDurationSeconds(state.String(), duration.Seconds())

	status := "success"
	if res.Error != nil {
		status = "error"
	}

	f.metrics.IncTGRequestsTotal(state.String(), status)

	return res
}
