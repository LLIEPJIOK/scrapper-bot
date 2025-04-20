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
	client   Client
	channels Channels
	cache    Cache
}

func NewTrackSaver(client Client, channels Channels, cache Cache) *TrackSaver {
	return &TrackSaver{
		client:   client,
		channels: channels,
		cache:    cache,
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
		state.ShowError = "ошибка при добавлении ссылки"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
			Error:            fmt.Errorf("client.AddLink(ctx, link): %w", err),
		}
	}

	if err := h.cache.InvalidateListLinks(ctx, state.ChatID); err != nil {
		slog.Error(
			"failed to invalidate links in cache",
			slog.Any("error", err),
			slog.Int64("chat_id", state.ChatID),
		)
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
