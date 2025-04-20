package processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TrackAddTimeSetterDigest struct {
	channels Channels
}

func NewTrackAddTimeSetterDigest(channels Channels) *TrackAddTimeSetterDigest {
	return &TrackAddTimeSetterDigest{
		channels: channels,
	}
}

func (h *TrackAddTimeSetterDigest) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	return setNotificationTime(ctx, state, false, h.channels.TelegramResp())
}

func setNotificationTime(
	_ context.Context,
	state *State,
	sendImmediately bool,
	ch chan<- tgbotapi.Chattable,
) *fsm.Result[*State] {
	link, ok := state.Object.(*domain.Link)
	if !ok {
		slog.Error(
			"invalid object type",
			slog.Any("type", fmt.Sprintf("%T", state.Object)),
			slog.Any("handler", "set notification time"),
			slog.Any("send_immediately", sendImmediately),
		)

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
		}
	}

	link.SendImmediately = domain.NewNull(sendImmediately)
	state.Object = link

	ans := "Можете настроить ссылку или сохранить её в текущем состоянии."
	msg := tgbotapi.NewEditMessageText(state.ChatID, state.MessageID, ans)

	keyboard := createKeyboard(link)
	if len(keyboard.InlineKeyboard) == 1 {
		return &fsm.Result[*State]{
			NextState:        trackSave,
			IsAutoTransition: true,
			Result:           state,
		}
	}

	msg.ReplyMarkup = &keyboard
	ch <- msg

	return &fsm.Result[*State]{
		NextState:        callback,
		IsAutoTransition: false,
		Result:           state,
	}
}
