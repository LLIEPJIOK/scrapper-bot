package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Lister struct {
	channels Channels
}

func NewLister(channels Channels) *Lister {
	return &Lister{
		channels: channels,
	}
}

func (h *Lister) Handle(_ context.Context, state *State) *fsm.Result[*State] {
	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Все", listAll.String()),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("С определённым тегом", listByTagInput.String()),
		},
	}

	msg := tgbotapi.NewMessage(state.ChatID, "Какие ссылки вы хотите получить?")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		NextState:        callback,
		IsAutoTransition: false,
		Result:           state,
	}
}
