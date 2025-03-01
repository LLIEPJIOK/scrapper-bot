package processor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TrackLinkAdder struct {
	validLinks []string
	channels   Channels
}

func NewTrackLinkAdder(channels Channels) *TrackLinkAdder {
	return &TrackLinkAdder{
		validLinks: []string{"https://stackoverflow.com/questions/", "https://github.com/"},
		channels:   channels,
	}
}

func (h *TrackLinkAdder) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	if !h.isValidLink(state.Message) {
		ans := "Неверный формат ссылки. Используйте следующие форматы:\n- "
		ans += strings.Join(h.validLinks, "\n -")

		msg := tgbotapi.NewMessage(state.ChatID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			NextState:        state.FSMState,
			IsAutoTransition: false,
			Result:           state,
		}
	}

	update := func(link *domain.Link, value string) *domain.Link {
		link.URL = value

		return link
	}

	return updateField(ctx, state, h.channels.TelegramResp(), update)
}

func (h *TrackLinkAdder) isValidLink(link string) bool {
	for _, validLink := range h.validLinks {
		if strings.HasPrefix(link, validLink) {
			return true
		}
	}

	return false
}

func createKeyboard(link *domain.Link) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0)

	if link.Tags == nil {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Добавить тэги", trackAddTags.String()),
		})
	}

	if link.Filters == nil {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Добавить фильтры", trackAddFilters.String()),
		})
	}

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Сохранить", trackSave.String()),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func updateField(
	ctx context.Context,
	state *State,
	ch chan tgbotapi.Chattable,
	update func(*domain.Link, string) *domain.Link,
) *fsm.Result[*State] {
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

	link = update(link, state.Message)
	state.Object = link

	ans := "Можете добавить опциональные поля или сохранить ссылку в текущем состоянии."
	msg := tgbotapi.NewMessage(state.ChatID, ans)

	keyboard := createKeyboard(link)
	if len(keyboard.InlineKeyboard) == 1 {
		return &fsm.Result[*State]{
			NextState:        trackSave,
			IsAutoTransition: true,
			Result:           state,
		}
	}

	msg.ReplyMarkup = keyboard
	ch <- msg

	return &fsm.Result[*State]{
		NextState:        callback,
		IsAutoTransition: false,
		Result:           state,
	}
}
