package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/cache/bot"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AllLister struct {
	client   Client
	channels Channels
	cache    Cache
}

func NewAllLister(client Client, channels Channels, cache Cache) *AllLister {
	return &AllLister{
		client:   client,
		channels: channels,
		cache:    cache,
	}
}

func (h *AllLister) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	list, err := h.cache.GetListLinks(ctx, state.ChatID, "")
	if err == nil {
		return h.sendList(ctx, state, list)
	}

	if errors.As(err, &bot.ErrNoData{}) {
		list, err = h.getLinks(ctx, state.ChatID)
	}

	if err != nil {
		state.ShowError = "не удалось получить ссылки"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
			Error:            err,
		}
	}

	if err := h.cache.SetListLinks(ctx, state.ChatID, "", list); err != nil {
		slog.Error(
			"failed to set list links",
			slog.Int64("chat_id", state.ChatID),
			slog.Any("error", err),
		)
	}

	return h.sendList(ctx, state, list)
}

func (h *AllLister) getLinks(ctx context.Context, chatID int64) (string, error) {
	links, err := h.client.GetLinks(ctx, chatID, "")
	if err != nil {
		return "", fmt.Errorf("h.client.GetLinks(ctx, %d, \"\"): %w", chatID, err)
	}

	if len(links) == 0 {
		return "У вас нет ни одной ссылки. Для добавления ссылки воспользуйтесь командой /track", nil
	}

	ansBuilder := strings.Builder{}
	ansBuilder.WriteString("Ваши ссылки:\n")

	for i, link := range links {
		ansBuilder.WriteString(fmt.Sprintf("%d) %s\n", i+1, link.URL))

		if len(link.Filters) > 0 {
			ansBuilder.WriteString(fmt.Sprintf("*Фильтры:* %s\n", strings.Join(link.Filters, "; ")))
		}

		if len(link.Tags) > 0 {
			ansBuilder.WriteString(fmt.Sprintf("#%s\n", strings.Join(link.Tags, " #")))
		}

		ansBuilder.WriteString("*Время отправки:* ")

		if link.SendImmediately.Value {
			ansBuilder.WriteString("сразу\n")
		} else {
			ansBuilder.WriteString("по расписанию\n")
		}

		ansBuilder.WriteString("\n")
	}

	return ansBuilder.String(), nil
}

func (h *AllLister) sendList(_ context.Context, state *State, list string) *fsm.Result[*State] {
	msg := tgbotapi.NewEditMessageText(state.ChatID, state.MessageID, list)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: false,
		Result:           state,
	}
}
