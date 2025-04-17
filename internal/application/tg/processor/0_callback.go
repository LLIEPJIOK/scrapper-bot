package processor

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Callbacker struct {
	channels Channels
}

func NewCallbacker(channels Channels) *Callbacker {
	return &Callbacker{
		channels: channels,
	}
}

func (h *Callbacker) Handle(_ context.Context, state *State) *fsm.Result[*State] {
	switch {
	case state.Message == trackAddTags.String():
		ans := "Введите теги через пробел."
		msg := tgbotapi.NewEditMessageText(state.ChatID, state.MessageID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			NextState:        trackAddTags,
			IsAutoTransition: false,
			Result:           state,
		}

	case state.Message == trackAddFilters.String():
		ans := "Введите фильтры через пробел.\nПример: user:dummy type:comment"
		msg := tgbotapi.NewEditMessageText(state.ChatID, state.MessageID, ans)
		h.channels.TelegramResp() <- msg

		return &fsm.Result[*State]{
			NextState:        trackAddFilters,
			IsAutoTransition: false,
			Result:           state,
		}

	case state.Message == trackSave.String():
		return &fsm.Result[*State]{
			NextState:        trackSave,
			IsAutoTransition: true,
			Result:           state,
		}

	case state.Message == listAll.String():
		return &fsm.Result[*State]{
			NextState:        listAll,
			IsAutoTransition: true,
			Result:           state,
		}

	case state.Message == listByTagInput.String():
		return &fsm.Result[*State]{
			NextState:        listByTagInput,
			IsAutoTransition: true,
			Result:           state,
		}

	default:
		state.ShowError = "неопознанная команда"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
		}
	}
}
