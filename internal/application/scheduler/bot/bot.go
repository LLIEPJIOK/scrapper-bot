package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/go-co-op/gocron/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Repository interface {
	GetUpdatesChats(ctx context.Context, from, to time.Time) ([]int64, error)
	GetUpdates(
		ctx context.Context,
		chatID int64,
		from, to time.Time,
	) ([]domain.Update, error)
}

type Channels interface {
	TelegramResp() chan tgbotapi.Chattable
}

type Scheduler struct {
	repo       Repository
	channels   Channels
	atHours    uint
	atMinutes  uint
	atSeconds  uint
	lastSended time.Time
}

func NewScheduler(cfg *config.Scheduler, repo Repository, channels Channels) *Scheduler {
	return &Scheduler{
		repo:       repo,
		channels:   channels,
		atHours:    cfg.AtHours,
		atMinutes:  cfg.AtMinutes,
		atSeconds:  cfg.AtSeconds,
		lastSended: time.Now().Add(-time.Hour),
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
			s.SendUpdates(ctx)
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

func (s *Scheduler) SendUpdates(ctx context.Context) {
	started := time.Now()

	chats, err := s.repo.GetUpdatesChats(ctx, s.lastSended, started)
	if err != nil {
		slog.Error("failed to get updates chats", slog.Any("error", err))

		return
	}

	for _, chat := range chats {
		updates, err := s.repo.GetUpdates(ctx, chat, s.lastSended, started)
		if err != nil {
			slog.Error("failed to get updates", slog.Any("error", err))

			return
		}

		msg := tgbotapi.NewMessage(chat, updatesToText(updates))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = true
		s.channels.TelegramResp() <- msg
	}

	s.lastSended = started
}

func updatesToText(updates []domain.Update) string {
	builder := strings.Builder{}
	builder.WriteString("Обновления по вашим ссылкам:\n\n")

	for i, update := range updates {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, update.Message))
	}

	return builder.String()
}
