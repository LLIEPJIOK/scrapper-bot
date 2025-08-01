package processor

import (
	"context"
	"log/slog"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const numWorkers = 10

type Client interface {
	RegisterChat(ctx context.Context, id int64) error
	AddLink(ctx context.Context, link *domain.Link) error
	DeleteLink(ctx context.Context, chatID int64, linkURL string) error
	GetLinks(ctx context.Context, chatID int64, tag string) ([]*domain.Link, error)
}

type Channels interface {
	TelegramReq() chan domain.TelegramRequest
	TelegramResp() chan tgbotapi.Chattable
}

type Metrics interface {
	IncTGRequestsTotal(state, status string)
	ObserveTGRequestsDurationSeconds(state string, seconds float64)
}

type Cache interface {
	GetListLinks(
		ctx context.Context,
		chatID int64,
		tag string,
	) (string, error)
	SetListLinks(
		ctx context.Context,
		chatID int64,
		tag string,
		list string,
	) error
	InvalidateListLinks(ctx context.Context, chatID int64) error
}

type Processor struct {
	fsm      *CustomFSM
	client   Client
	channels Channels
	mu       sync.RWMutex
	states   map[int64]*State
}

func New(client Client, channels Channels, cache Cache, metrics Metrics) *Processor {
	fsmBuilder := fsm.NewBuilder[*State]()
	fsmBuilder.
		AddState(callback, NewCallbacker(channels)).
		AddState(command, NewCommander()).
		AddState(start, NewStater(client, channels)).
		AddState(help, NewHelper(channels)).
		AddState(track, NewTracker(channels)).
		AddState(trackAddLink, NewTrackLinkAdder(client, channels)).
		AddState(trackAddFilters, NewTrackFilterAdder(channels)).
		AddState(trackAddTags, NewTrackTagAdder(channels)).
		AddState(trackAddSetTime, NewTrackAddTimeSetter(channels)).
		AddState(trackAddSetTimeDigest, NewTrackAddTimeSetterDigest(channels)).
		AddState(trackAddSetTimeImmediately, NewTrackAddTimeSetterImmediately(channels)).
		AddState(trackSave, NewTrackSaver(client, channels, cache)).
		AddState(list, NewLister(channels)).
		AddState(listAll, NewAllLister(client, channels, cache)).
		AddState(listByTagInput, NewByTagInputLister(channels)).
		AddState(listByTag, NewByTagLister(client, channels, cache)).
		AddState(untrack, NewUntracker(channels)).
		AddState(untrackDeleteLink, NewUntrackLinkDeleter(client, channels, cache)).
		AddState(fail, NewFailer(channels)).
		AddTransition(callback, trackAddTags).
		AddTransition(callback, trackAddFilters).
		AddTransition(callback, trackAddSetTime).
		AddTransition(callback, trackAddSetTimeDigest).
		AddTransition(callback, trackAddSetTimeImmediately).
		AddTransition(callback, trackSave).
		AddTransition(callback, listAll).
		AddTransition(callback, listByTagInput).
		AddTransition(callback, fail).
		AddTransition(command, start).
		AddTransition(command, help).
		AddTransition(command, track).
		AddTransition(command, list).
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
		AddTransition(trackAddSetTimeDigest, trackSave).
		AddTransition(trackAddSetTimeDigest, callback).
		AddTransition(trackAddSetTimeImmediately, trackSave).
		AddTransition(trackAddSetTimeImmediately, callback).
		AddTransition(trackSave, fail).
		AddTransition(list, fail).
		AddTransition(listAll, fail).
		AddTransition(listByTag, fail).
		AddTransition(untrack, untrackDeleteLink).
		AddTransition(untrackDeleteLink, fail)

	return &Processor{
		client:   client,
		channels: channels,
		fsm:      NewCustomFSM(fsmBuilder.Build(), metrics),
		mu:       sync.RWMutex{},
		states:   make(map[int64]*State),
	}
}

func (p *Processor) GetState(chatID int64) (*State, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	state, ok := p.states[chatID]

	return state, ok
}

func (p *Processor) SetState(chatID int64, state *State) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.states[chatID] = state
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

				state, ok := p.GetState(req.ChatID)
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

				state, ok := p.GetState(req.ChatID)
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
