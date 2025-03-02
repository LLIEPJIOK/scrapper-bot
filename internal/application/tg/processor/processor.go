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
	DeleteLink(ctx context.Context, chatID int64, linkURL string) error
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
	states   map[int64]*State
}

func New(client Client, channels Channels) *Processor {
	fsmBuilder := fsm.NewBuilder[*State]()
	fsmBuilder.
		AddState(callback, NewCallbacker(channels)).
		AddState(command, NewCommander()).
		AddState(start, NewStater(client, channels)).
		AddState(help, NewHelper(channels)).
		AddState(track, NewTracker(channels)).
		AddState(trackAddLink, NewTrackLinkAdder(channels)).
		AddState(trackAddFilters, NewTrackFilterAdder(channels)).
		AddState(trackAddTags, NewTrackTagAdder(channels)).
		AddState(trackSave, NewTrackSaver(client, channels)).
		AddState(trackList, NewTrackLister(client, channels)).
		AddState(untrack, NewUntracker(channels)).
		AddState(untrackDeleteLink, NewUntrackLinkDeleter(client, channels)).
		AddState(fail, NewFailer(channels)).
		AddTransition(callback, trackAddTags).
		AddTransition(callback, trackAddFilters).
		AddTransition(callback, trackSave).
		AddTransition(callback, fail).
		AddTransition(command, start).
		AddTransition(command, help).
		AddTransition(command, track).
		AddTransition(command, trackList).
		AddTransition(command, untrack).
		AddTransition(command, fail).
		AddTransition(start, fail).
		AddTransition(track, trackAddLink).
		AddTransition(trackAddLink, fail).
		AddTransition(trackAddLink, trackSave).
		AddTransition(trackAddLink, callback).
		AddTransition(trackAddFilters, trackSave).
		AddTransition(trackAddFilters, callback).
		AddTransition(trackAddTags, trackSave).
		AddTransition(trackAddTags, callback).
		AddTransition(trackSave, fail).
		AddTransition(trackList, fail).
		AddTransition(untrack, untrackDeleteLink).
		AddTransition(untrackDeleteLink, fail)

	return &Processor{
		client:   client,
		channels: channels,
		fsm:      fsmBuilder.Build(),
		states:   make(map[int64]*State),
	}
}

func (p *Processor) Run(ctx context.Context) error {
	workCh := make(chan *State, numWorkers)
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
				slog.Info("getting message", slog.Any("message", req.Message))

				state, ok := p.states[req.ChatID]
				if !ok || state.FSMState.String() == "" {
					workCh <- &State{
						FSMState:  fail,
						ChatID:    req.ChatID,
						ShowError: "неопознанная команда",
					}

					continue
				}

				state.Message = req.Message
				state.MessageID = 0
				workCh <- state

			case domain.Command:
				slog.Info("getting command", slog.Any("command", req.Message))

				workCh <- &State{
					FSMState: command,
					ChatID:   req.ChatID,
					Message:  req.Message,
				}

			case domain.Callback:
				slog.Info("getting callback", slog.Any("callback", req.Message))

				state, ok := p.states[req.ChatID]
				if !ok || state.FSMState != callback {
					workCh <- &State{
						FSMState:  fail,
						ChatID:    req.ChatID,
						MessageID: req.MessageID,
						ShowError: "неопознанная команда",
					}

					continue
				}

				state.Message = req.Message
				state.MessageID = req.MessageID
				workCh <- state

			default:
				slog.Warn("unknown request type", slog.Any("type", req.Type))
				continue
			}
		}
	}
}
