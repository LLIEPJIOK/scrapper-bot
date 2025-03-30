package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ByTagInputLister struct {
	channels Channels
}

func NewByTagInputLister(channels Channels) *ByTagInputLister {
	return &ByTagInputLister{
		channels: channels,
	}
}

func (h *ByTagInputLister) Handle(_ context.Context, state *State) *fsm.Result[*State] {
	msg := tgbotapi.NewEditMessageText(state.ChatID, state.MessageID, "Введите тег")
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		NextState:        listByTag,
		IsAutoTransition: false,
		Result:           state,
	}
}
