package processor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/http/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/cache/bot"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ByTagLister struct {
	client   Client
	channels Channels
	cache    Cache
}

func NewByTagLister(client Client, channels Channels, cache Cache) *ByTagLister {
	return &ByTagLister{
		client:   client,
		channels: channels,
		cache:    cache,
	}
}

func (h *ByTagLister) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	state.Message = strings.TrimSpace(state.Message)

	list, err := h.cache.GetListLinks(ctx, state.ChatID, state.Message)
	if err == nil {
		return h.sendList(ctx, state, list)
	}

	if !errors.As(err, &bot.ErrNoData{}) {
		return h.handleError(state, err)
	}

	links, err := h.client.GetLinks(ctx, state.ChatID, state.Message)

	userErr := &scrapper.ErrUserResponse{}

	if errors.As(err, userErr) {
		ans := userErr.Message
		msg := tgbotapi.NewMessage(state.ChatID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			NextState:        state.FSMState,
			IsAutoTransition: false,
			Result:           state,
		}
	}

	if err != nil {
		return h.handleError(state, err)
	}

	list = h.linksToText(state, links)

	if err := h.cache.SetListLinks(ctx, state.ChatID, state.Message, list); err != nil {
		slog.Error(
			"failed to set list links",
			slog.Int64("chat_id", state.ChatID),
			slog.String("tag", state.Message),
			slog.Any("error", err),
		)
	}

	return h.sendList(ctx, state, list)
}

func (h *ByTagLister) linksToText(state *State, links []*domain.Link) string {
	if len(links) == 0 {
		return fmt.Sprintf(
			"У вас нет ссылок с тегом #%s. Для добавления ссылки воспользуйтесь командой /track",
			state.Message,
		)
	}

	ansBuilder := strings.Builder{}
	ansBuilder.WriteString(fmt.Sprintf("Ваши ссылки c тегом #%s:\n", state.Message))

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

	return ansBuilder.String()
}

func (h *ByTagLister) sendList(_ context.Context, state *State, list string) *fsm.Result[*State] {
	msg := tgbotapi.NewMessage(state.ChatID, list)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: false,
		Result:           state,
	}
}

func (h *ByTagLister) handleError(state *State, err error) *fsm.Result[*State] {
	state.ShowError = "не удалось получить ссылки"

	return &fsm.Result[*State]{
		NextState:        fail,
		IsAutoTransition: true,
		Result:           state,
		Error:            err,
	}
}
