package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Failer struct {
	fsm.BaseTransition
	channels Channels
}

func NewFailer(channels Channels) *Failer {
	return &Failer{
		BaseTransition: fsm.BaseTransition{
			Auto: true,
		},
		channels: channels,
	}
}

func (f *Failer) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	msgString := "Не удалось обработать запрос"
	if state.ShowError != "" {
		msgString += ": " + state.ShowError
	}

	msgString += ". Попробуйте повторить запрос позже."
	msg := tgbotapi.NewMessage(state.ChatID, msgString)
	f.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		Result: state,
	}
}
