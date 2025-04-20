package scrapper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/go-co-op/gocron/v2"
)

const (
	numWorkers = 5
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
	GetUpdates(link string, from, to time.Time) ([]string, error)
}

type Client interface {
	UpdatesPost(ctx context.Context, update *domain.Update) error
}

type Scheduler struct {
	repo     Repository
	client   Client
	checkers []Checher
	interval time.Duration
	pageSize uint
}

func NewScheduler(
	cfg *config.ScrapperScheduler,
	repo Repository,
	client Client,
	checkers ...Checher,
) *Scheduler {
	return &Scheduler{
		repo:     repo,
		client:   client,
		checkers: checkers,
		interval: cfg.Interval,
		pageSize: cfg.PageSize,
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
		links, err := s.repo.GetCheckLinks(ctx, cursor, started, s.pageSize)
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

	var (
		updates []string
		err     error
	)

	if len(link.Chats) != 0 {
		for _, checker := range s.checkers {
			updates, err = checker.GetUpdates(link.URL, link.CheckedAt, tm)
			if err != nil {
				slog.Error(
					"failed to get updates",
					slog.Any("url", link.URL),
					slog.Any("error", err),
				)

				continue
			}

			if len(updates) != 0 {
				break
			}
		}
	}

	s.sendUpdates(ctx, link, updates)

	if len(updates) != 0 || err == nil {
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

func (s *Scheduler) sendUpdates(ctx context.Context, link *domain.CheckLink, updates []string) {
	for _, update := range updates {
		for _, chat := range link.Chats {
			if !isValidUpdate(update, chat.Filters) {
				continue
			}

			err := s.client.UpdatesPost(ctx, &domain.Update{
				ChatID:  chat.ChatID,
				URL:     link.URL,
				Message: update,
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
}

func isValidUpdate(update string, filters []string) bool {
	for _, filter := range filters {
		switch {
		case strings.HasPrefix(filter, "user="):
			if strings.Contains(
				update,
				fmt.Sprintf("<b>Автор</b>: <i>%s</i>\n", strings.TrimPrefix(filter, "user=")),
			) {
				return false
			}
		}
	}

	return true
}
