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

const numWorkers = 5

type Repository interface {
	GetCheckLinks() []*domain.CheckLink
	UpdateCheckTime(url string, checkedAt time.Time) error
}

type Checher interface {
	HasUpdates(link string, lastUpdate time.Time) (bool, error)
}

type Client interface {
	UpdatesPost(ctx context.Context, link string, chats []int64) error
}

type ScrapperScheduler struct {
	repo     Repository
	client   Client
	checkers []Checher
	interval time.Duration
}

func NewScrapperScheduler(
	cfg *config.Scheduler,
	repo Repository,
	client Client,
	checkers ...Checher,
) *ScrapperScheduler {
	return &ScrapperScheduler{
		repo:     repo,
		client:   client,
		checkers: checkers,
		interval: cfg.Interval,
	}
}

func (s *ScrapperScheduler) Run(ctx context.Context) error {
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
			links := s.repo.GetCheckLinks()
			for _, link := range links {
				ch <- link
			}
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

func (s *ScrapperScheduler) worker(ctx context.Context, ch <-chan *domain.CheckLink) {
	for {
		select {
		case <-ctx.Done():
			return

		case link := <-ch:
			tm := time.Now()
			hasUpdates, hasError := false, false

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

			if hasUpdates {
				err := s.client.UpdatesPost(ctx, link.URL, link.Chats)
				if err != nil {
					slog.Error(
						"failed to send updates",
						slog.Any("url", link.URL),
						slog.Any("chats", link.Chats),
						slog.Any("error", err),
					)
				}
			}

			if hasUpdates || !hasError {
				err := s.repo.UpdateCheckTime(link.URL, tm)
				if err != nil {
					slog.Error(
						"failed to update check time",
						slog.Any("url", link.URL),
						slog.Any("error", err),
					)
				}
			}
		}
	}
}
