package processor

import (
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

const (
	callback fsm.State = "callback"
	command  fsm.State = "command"

	start fsm.State = "start"

	help fsm.State = "help"

	trackList fsm.State = "list"

	track           fsm.State = "track"
	trackAddLink    fsm.State = "track_add_link"
	trackAddFilters fsm.State = "track_add_filters"
	trackAddTags    fsm.State = "track_add_tags"
	trackSave       fsm.State = "track_save"

	untrack           fsm.State = "untrack"
	untrackDeleteLink fsm.State = "untrack_delete_link"

	fail fsm.State = "fail"
)

type State struct {
	FSMState  fsm.State
	ChatID    int64
	MessageID int
	Message   string
	Object    any
	ShowError string
}
