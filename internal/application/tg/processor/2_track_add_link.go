package processor

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TrackLinkAdder struct {
	client   Client
	channels Channels
	formats  []string
	regexes  []*regexp.Regexp
}

func NewTrackLinkAdder(client Client, channels Channels) *TrackLinkAdder {
	return &TrackLinkAdder{
		client:   client,
		channels: channels,
		formats: []string{
			"https://stackoverflow.com/questions/{id}/{title}",
			"https://github.com/{user}/{repo}",
			"https://github.com/{user}/{repo}/issues/{id}",
			"https://github.com/{user}/pull/{id}",
		},
		regexes: []*regexp.Regexp{
			regexp.MustCompile(`^https://stackoverflow\.com/questions/(\d+)/([\w-]+)$`),
			regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)$`),
			regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)/issues/(\d+)$`),
			regexp.MustCompile(`^https://github\.com/([\w.-]+)/([\w.-]+)/pull/(\d+)$`),
		},
	}
}

func (h *TrackLinkAdder) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	if !h.isValidLink(state.Message) {
		ans := "Неверный формат ссылки. Используйте следующие форматы:\n- "
		ans += strings.Join(h.formats, "\n- ")

		msg := tgbotapi.NewMessage(state.ChatID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			NextState:        state.FSMState,
			IsAutoTransition: false,
			Result:           state,
		}
	}

	exists, err := h.IsLinkExists(ctx, state.Message, state.ChatID)
	if err != nil {
		state.ShowError = "ошибка при добавлении ссылки"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
			Error: fmt.Errorf(
				"h.IsLinkExists(ctx, %q, %d): %w",
				state.Message,
				state.ChatID,
				err,
			),
		}
	}

	if exists {
		ans := "Ссылка уже существует. Введите другую ссылку или посмотрите список, используя /list"

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
	for _, reges := range h.regexes {
		if reges.MatchString(link) {
			return true
		}
	}

	return false
}

func (h *TrackLinkAdder) IsLinkExists(
	ctx context.Context,
	url string,
	chatID int64,
) (bool, error) {
	links, err := h.client.GetLinks(ctx, chatID)
	if err != nil {
		return false, fmt.Errorf("failed to get links: %w", err)
	}

	exists := slices.ContainsFunc(links, func(link *domain.Link) bool {
		return link.URL == url
	})

	return exists, nil
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
	_ context.Context,
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
