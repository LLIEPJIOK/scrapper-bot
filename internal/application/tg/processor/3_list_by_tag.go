package processor

import (
	"context"
	"fmt"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ByTagLister struct {
	client   Client
	channels Channels
}

func NewByTagLister(client Client, channels Channels) *ByTagLister {
	return &ByTagLister{
		client:   client,
		channels: channels,
	}
}

func (h *ByTagLister) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	state.Message = strings.TrimSpace(state.Message)

	links, err := h.client.GetLinks(ctx, state.ChatID, state.Message)
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
		ans := fmt.Sprintf(
			"У вас нет ссылок с тегом #%s. Для добавления ссылки воспользуйтесь командой /track",
			state.Message,
		)
		msg := tgbotapi.NewMessage(state.ChatID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			IsAutoTransition: false,
			Result:           state,
		}
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

		ansBuilder.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(state.ChatID, ansBuilder.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: false,
		Result:           state,
	}
}
