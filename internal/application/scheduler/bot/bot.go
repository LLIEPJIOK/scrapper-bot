package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	repository "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/bot"
	"github.com/go-co-op/gocron/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Repository interface {
	GetUpdates() ([]*repository.UpdateChat, error)
}

type Channels interface {
	TelegramResp() chan tgbotapi.Chattable
}

type Scheduler struct {
	repo      Repository
	channels  Channels
	atHours   uint
	atMinutes uint
	atSeconds uint
}

func NewScheduler(cfg *config.Scheduler, repo Repository, channels Channels) *Scheduler {
	return &Scheduler{
		repo:      repo,
		channels:  channels,
		atHours:   cfg.AtHours,
		atMinutes: cfg.AtMinutes,
		atSeconds: cfg.AtSeconds,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	schedule, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	_, err = schedule.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(gocron.NewAtTime(s.atHours, s.atMinutes, s.atSeconds)),
		),
		gocron.NewTask(func() {
			updates, err := s.repo.GetUpdates()
			if err != nil {
				slog.Error("failed to get updates", slog.Any("error", err))

				return
			}

			for _, update := range updates {
				ans := updateToText(update)
				msg := tgbotapi.NewMessage(update.ID, ans)
				s.channels.TelegramResp() <- msg
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

func updateToText(update *repository.UpdateChat) string {
	builder := strings.Builder{}
	builder.WriteString("Обновления по вашим ссылкам:\n")

	for i, link := range update.Links {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, link))
	}

	return builder.String()
}
