package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

type TrackAddTimeSetterImmediately struct {
	channels Channels
}

func NewTrackAddTimeSetterImmediately(channels Channels) *TrackAddTimeSetterImmediately {
	return &TrackAddTimeSetterImmediately{
		channels: channels,
	}
}

func (h *TrackAddTimeSetterImmediately) Handle(
	ctx context.Context,
	state *State,
) *fsm.Result[*State] {
	return setNotificationTime(ctx, state, true, h.channels.TelegramResp())
}
