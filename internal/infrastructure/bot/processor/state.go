package processor

import (
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

const (
	command fsm.State = "command"
	start   fsm.State = "start"
	fail    fsm.State = "fail"
)

type State struct {
	FSMState  fsm.State
	ChatID    int64
	Message   string
	ShowError string
}
