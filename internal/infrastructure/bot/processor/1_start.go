package processor

import (
	"context"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const staterAnswer = `*Привет! Я LinkTracker – твой помощник для отслеживания обновлений на сайтах.*  

🔹 Подписывайся на ссылки и получай уведомления об изменениях.  
🔹 Управляй подписками прямо в Telegram.  
🔹 Получай обновления сразу или в удобное время.  

📌 Доступные команды:  
- /track – подписаться на обновления
- /untrack – отписаться от обновлений
- /list – показать все подписки
- /help – справка по командам

Начни с /track и будь в курсе важных событий! 🚀
`

type Stater struct {
	client   Client
	channels Channels
}

func NewStater(client Client, channels Channels) *Stater {
	return &Stater{
		client:   client,
		channels: channels,
	}
}

func (h *Stater) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	if err := h.client.RegisterChat(ctx, state.ChatID); err != nil {
		state.ShowError = "ошибка регистрации чата"

		return &fsm.Result[*State]{
			NextState:        fail,
			IsAutoTransition: true,
			Result:           state,
			Error:            fmt.Errorf("h.client.RegisterChat(ctx, %d): %w", state.ChatID, err),
		}
	}

	msg := tgbotapi.NewMessage(state.ChatID, staterAnswer)
	msg.ParseMode = tgbotapi.ModeMarkdown
	h.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		IsAutoTransition: false,
		Result:           state,
	}
}
