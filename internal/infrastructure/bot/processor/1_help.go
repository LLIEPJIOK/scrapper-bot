package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const helperAnswer = `📌 Доступные команды:  
- /track – подписаться на обновления
- /track – отписаться от обновлений
- /list – показать все подписки
- /help – справка по командам
`

type Helper struct {
	channels Channels
}

func NewHelper(channels Channels) *Helper {
	return &Helper{
		channels: channels,
	}
}

func (h *Helper) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	msg := tgbotapi.NewMessage(state.ChatID, helperAnswer)
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: false,
		Result:           state,
	}
}
