package scrapper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/go-co-op/gocron/v2"
)

const (
	numWorkers     = 5
	paginationSize = 100
)

type Repository interface {
	GetCheckLinks(
		ctx context.Context,
		from, to time.Time,
		limit uint,
	) ([]*domain.CheckLink, error)
	UpdateCheckTime(ctx context.Context, url string, checkedAt time.Time) error
}

type Checher interface {
	HasUpdates(link string, lastUpdate time.Time) (bool, error)
}

type Client interface {
	UpdatesPost(ctx context.Context, update *domain.Update) error
}

type Scheduler struct {
	repo     Repository
	client   Client
	checkers []Checher
	interval time.Duration
}

func NewScheduler(
	cfg *config.Scheduler,
	repo Repository,
	client Client,
	checkers ...Checher,
) *Scheduler {
	return &Scheduler{
		repo:     repo,
		client:   client,
		checkers: checkers,
		interval: cfg.Interval,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	ch := make(chan *domain.CheckLink)

	for range numWorkers {
		go s.worker(ctx, ch)
	}

	schedule, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	_, err = schedule.NewJob(
		gocron.DurationJob(s.interval),
		gocron.NewTask(func() {
			s.checker(ctx, ch)
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create scheduler job: %w", err)
	}

	schedule.Start()

	<-ctx.Done()

	err = schedule.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown scheduler: %w", err)
	}

	return nil
}

func (s *Scheduler) checker(ctx context.Context, ch chan<- *domain.CheckLink) {
	started := time.Now()
	cursor := time.Time{}

	for {
		links, err := s.repo.GetCheckLinks(ctx, cursor, started, paginationSize)
		if err != nil {
			slog.Error("failed to get links", "err", err)

			return
		}

		if len(links) == 0 {
			break
		}

		cursor = links[len(links)-1].CheckedAt

		for _, link := range links {
			ch <- link
		}
	}
}

func (s *Scheduler) worker(ctx context.Context, ch <-chan *domain.CheckLink) {
	for {
		select {
		case <-ctx.Done():
			return

		case link := <-ch:
			s.getUpdates(ctx, link)
		}
	}
}

func (s *Scheduler) getUpdates(ctx context.Context, link *domain.CheckLink) {
	tm := time.Now()
	hasUpdates, hasError := false, false

	if len(link.Chats) != 0 {
		for _, checker := range s.checkers {
			has, err := checker.HasUpdates(link.URL, link.CheckedAt)
			if err != nil {
				slog.Error(
					"failed to check updates",
					slog.Any("url", link.URL),
					slog.Any("error", err),
				)

				hasError = true

				continue
			}

			if has {
				hasUpdates = true

				break
			}
		}
	}

	if hasUpdates {
		for _, chat := range link.Chats {
			err := s.client.UpdatesPost(ctx, &domain.Update{
				ChatID:  chat.ChatID,
				URL:     link.URL,
				Message: fmt.Sprintf("New updates on %s", link.URL),
				Tags:    chat.Tags,
			})
			if err != nil {
				slog.Error(
					"failed to send updates",
					slog.Any("url", link.URL),
					slog.Any("chats", link.Chats),
					slog.Any("error", err),
				)
			}
		}
	}

	if hasUpdates || !hasError {
		err := s.repo.UpdateCheckTime(ctx, link.URL, tm)
		if err != nil {
			slog.Error(
				"failed to update check time",
				slog.Any("url", link.URL),
				slog.Any("error", err),
			)
		}
	}
}
