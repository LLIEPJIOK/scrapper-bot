package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const TrackAddTimeSetterText = `Выберите способ отправки уведомлений:
1. **По расписанию**: уведомления будут отправляться каждый день в 10:00 UTC
2. **Сразу**: уведомления будут отправляться сразу после обнаружения изменений
`

type TrackAddTimeSetter struct {
	channels Channels
}

func NewTrackAddTimeSetter(channels Channels) *TrackAddTimeSetter {
	return &TrackAddTimeSetter{
		channels: channels,
	}
}

func (h *TrackAddTimeSetter) Handle(_ context.Context, state *State) *fsm.Result[*State] {
	rows := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("По расписанию", trackAddSetTimeDigest.String()),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("Сразу", trackAddSetTimeImmediately.String()),
		},
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewEditMessageText(state.ChatID, state.MessageID, TrackAddTimeSetterText)
	msg.ReplyMarkup = &markup
	msg.ParseMode = tgbotapi.ModeMarkdown

	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		NextState:        callback,
		IsAutoTransition: false,
		Result:           state,
	}
}
