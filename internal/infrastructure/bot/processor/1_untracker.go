package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Untracker struct {
	channels Channels
}

func NewUntracker(channels Channels) *Untracker {
	return &Untracker{
		channels: channels,
	}
}

func (h *Untracker) Handle(_ context.Context, state *State) *fsm.Result[*State] {
	ans := "Введите ссылку на ресурс, который хотите удалить."
	msg := tgbotapi.NewMessage(state.ChatID, ans)
	h.channels.TelegramResp() <- msg

	state.Object = &domain.Link{
		ChatID: state.ChatID,
	}

	return &fsm.Result[*State]{
		NextState:        untrackDeleteLink,
		IsAutoTransition: false,
		Result:           state,
	}
}
