package processor

import (
	"context"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

type TrackTagAdder struct {
	channels Channels
}

func NewTrackTagAdder(channels Channels) *TrackTagAdder {
	return &TrackTagAdder{
		channels: channels,
	}
}

func (h *TrackTagAdder) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	update := func(link *domain.Link, value string) *domain.Link {
		link.Tags = strings.Fields(state.Message)

		return link
	}

	return updateField(ctx, state, h.channels.TelegramResp(), update)
}
