package processor

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/scrapper/client"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UntrackLinkDeleter struct {
	client   Client
	channels Channels
}

func NewUntrackLinkDeleter(client Client, channels Channels) *UntrackLinkDeleter {
	return &UntrackLinkDeleter{
		client:   client,
		channels: channels,
	}
}

func (h *UntrackLinkDeleter) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	_, err := url.Parse(state.Message)
	if err != nil {
		ans := "Некорректный формат ссылки. Введите ссылку, которая отображается в /list"
		msg := tgbotapi.NewMessage(state.ChatID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			NextState:        state.FSMState,
			IsAutoTransition: false,
			Result:           state,
		}
	}

	err = h.client.DeleteLink(ctx, state.ChatID, state.Message)
	userErr := &client.ErrUserResponse{}
	if errors.Is(err, userErr) {
		ans := userErr.Message + ". Введите ссылку, которая отображается в /list"
		msg := tgbotapi.NewMessage(state.ChatID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			NextState:        state.FSMState,
			IsAutoTransition: false,
			Result:           state,
		}
	}

	if err != nil {
		state.ShowError = "ошибка при удалении ссылки"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
			Error: fmt.Errorf(
				"h.client.DeleteLink(ctx, %d, %q)",
				state.ChatID,
				state.Message,
			),
		}
	}

	ans := "Ссылка успешно удалена!"
	msg := tgbotapi.NewMessage(state.ChatID, ans)
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: false,
		Result:           state,
	}
}
