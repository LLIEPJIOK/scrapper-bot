package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Failer struct {
	channels Channels
}

func NewFailer(channels Channels) *Failer {
	return &Failer{
		channels: channels,
	}
}

func (h *Failer) Handle(_ context.Context, state *State) *fsm.Result[*State] {
	ans := "Не удалось обработать запрос"
	if state.ShowError != "" {
		ans += ": " + state.ShowError
	}

	ans += ". Попробуйте повторить запрос позже."

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
