package processor

import (
	"context"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
)

type TrackFilterAdder struct {
	channels Channels
}

func NewTrackFilterAdder(channels Channels) *TrackFilterAdder {
	return &TrackFilterAdder{
		channels: channels,
	}
}

func (h *TrackFilterAdder) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	update := func(link *domain.Link, value string) *domain.Link {
		link.Filters = strings.Fields(value)

		return link
	}

	return updateField(ctx, state, h.channels.TelegramResp(), update)
}
