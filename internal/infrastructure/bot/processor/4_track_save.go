package processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TrackSaver struct {
	channels Channels
	client   Client
}

func NewTrackSaver(channels Channels, client Client) *TrackSaver {
	return &TrackSaver{
		channels: channels,
		client:   client,
	}
}

func (h *TrackSaver) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	link, ok := state.Object.(*domain.Link)
	if !ok {
		slog.Error(
			"invalid object type",
			slog.Any("type", fmt.Sprintf("%T", state.Object)),
			slog.Any("handler", "TrackLinkAdder"),
		)

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
		}
	}

	state.Object = nil

	err := h.client.AddLink(ctx, link)
	if err != nil {
		state.ShowError = "не удалось добавить ссылку"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
			Error:            fmt.Errorf("client.AddLink(): %w", err),
		}
	}

	ans := "Ссылка успешно добавлена!"

	var msg tgbotapi.Chattable = tgbotapi.NewMessage(state.ChatID, ans)

	if state.MessageID != 0 {
		msg = tgbotapi.NewEditMessageText(state.ChatID, state.MessageID, ans)
	}

	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: false,
		Result:           state,
	}
}
