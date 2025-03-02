package processor

import (
	"context"
	"fmt"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TrackLister struct {
	client   Client
	channels Channels
}

func NewTrackLister(client Client, channels Channels) *TrackLister {
	return &TrackLister{
		client:   client,
		channels: channels,
	}
}

func (h *TrackLister) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	links, err := h.client.GetLinks(ctx, state.ChatID)
	if err != nil {
		state.ShowError = "не удалось получить ссылки"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
			Error:            fmt.Errorf("h.client.GetLinks(ctx, %d): %w", state.ChatID, err),
		}
	}

	if len(links) == 0 {
		ans := "У вас нет ни одной ссылки. Для добавления ссылки воспользуйтесь командой /track"
		msg := tgbotapi.NewMessage(state.ChatID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			IsAutoTransition: true,
			Result:           state,
		}
	}

	ansBuilder := strings.Builder{}
	ansBuilder.WriteString("Ваши ссылки:\n")

	for i, link := range links {
		ansBuilder.WriteString(fmt.Sprintf("%d) %s\n", i+1, link.URL))

		if len(link.Tags) > 0 {
			ansBuilder.WriteString(fmt.Sprintf("*Тэги:* %s\n", strings.Join(link.Tags, "; ")))
		}

		if len(link.Filters) > 0 {
			ansBuilder.WriteString(fmt.Sprintf("*Фильтры:* %s\n", strings.Join(link.Filters, "; ")))
		}

		ansBuilder.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(state.ChatID, ansBuilder.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: true,
		Result:           state,
	}
}
