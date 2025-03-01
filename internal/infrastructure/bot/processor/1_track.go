package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const trackerAnswer = `Введите ссылку на ресурс, который хотите отслеживать.
Доступные сайты:
	- GitHub
	- StackOverflow
`

type Tracker struct {
	channels Channels
}

func NewTracker(channels Channels) *Tracker {
	return &Tracker{
		channels: channels,
	}
}

func (h *Tracker) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	msg := tgbotapi.NewMessage(state.ChatID, trackerAnswer)
	h.channels.TelegramResp() <- msg

	state.Object = &domain.Link{
		ChatID: state.ChatID,
	}

	return &fsm.Result[*State]{
		NextState:        trackAddLink,
		IsAutoTransition: false,
		Result:           state,
	}
}
