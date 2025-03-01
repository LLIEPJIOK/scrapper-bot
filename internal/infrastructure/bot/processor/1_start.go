package processor

import (
	"context"
	"fmt"

	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const StaterAnswer = `**Привет! Я LinkTracker – твой помощник для отслеживания обновлений на сайтах.**  

🔹 Подписывайся на ссылки и получай уведомления об изменениях.  
🔹 Управляй подписками прямо в Telegram.  
🔹 Получай обновления сразу или в удобное время.  

📌 Доступные команды:  
- /track – подписаться на обновления  
- /untrack – отменить подписку  
- /list – показать все подписки  
- /help – справка по командам  

Начни с /track и будь в курсе важных событий! 🚀
`

type Stater struct {
	fsm.BaseTransition

	client   Client
	channels Channels
}

func NewStater(client Client, channels Channels) *Stater {
	return &Stater{
		BaseTransition: fsm.BaseTransition{
			Auto: true,
		},
		client:   client,
		channels: channels,
	}
}

func (s *Stater) Handle(ctx context.Context, state *State) *fsm.Result[*State] {
	if err := s.client.RegisterChat(ctx, state.ChatID); err != nil {
		state.ShowError = "ошибка регистрации чата"
		return &fsm.Result[*State]{
			NextState: fail,
			Result:    state,
			Error:     fmt.Errorf("failed to register chat: %w", err),
		}
	}

	msg := tgbotapi.NewMessage(state.ChatID, StaterAnswer)
	msg.ParseMode = "Markdown"
	s.channels.TelegramResp() <- msg

	return &fsm.Result[*State]{
		NextState: "none",
		Result:    state,
	}
}
