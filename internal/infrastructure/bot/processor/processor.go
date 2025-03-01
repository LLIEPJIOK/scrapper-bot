package processor

import (
	"context"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const numWorkers = 10

type Client interface {
	RegisterChat(ctx context.Context, id int64) error
	AddLink(ctx context.Context, link *domain.Link) error
	DeleteLink(ctx context.Context, link *domain.Link) error
	GetLinks(ctx context.Context, chatID int64) ([]*domain.Link, error)
}

type Channels interface {
	TelegramReq() chan domain.TelegramRequest
	TelegramResp() chan tgbotapi.Chattable
}

type Processor struct {
	fsm      *fsm.FSM[*State]
	client   Client
	channels Channels
}

func New(client Client, channels Channels) *Processor {
	fsmBuilder := fsm.NewBuilder[*State]()
	fsmBuilder.
		AddState(command, NewCommander()).
		AddState(start, NewStater(client, channels)).
		AddState(fail, NewFailer(channels)).
		AddTransition(command, start).
		AddTransition(start, fail)

	return &Processor{
		client:   client,
		channels: channels,
		fsm:      fsmBuilder.Build(),
	}
}

func (p *Processor) Run(ctx context.Context) error {
	workCh := make(chan State, numWorkers)
	defer close(workCh)

	for range numWorkers {
		go p.worker(ctx, workCh)
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case req := <-p.channels.TelegramReq():
			switch req.Type {
			case domain.Message:
			case domain.Command:
				slog.Info("getting command", slog.Any("command", req.Message))

				workCh <- State{
					FSMState: command,
					ChatID:   req.ChatID,
					Message:  req.Message,
				}

			case domain.Callback:
			default:
				slog.Warn("unknown request type", slog.Any("type", req.Type))
				continue
			}
		}
	}
}
