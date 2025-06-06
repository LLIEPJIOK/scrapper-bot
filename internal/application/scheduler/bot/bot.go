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
	GetUpdatesChats(ctx context.Context) ([]int64, error)
	GetAndClearUpdates(
		ctx context.Context,
		chatID int64,
	) ([]*domain.Update, error)
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

func NewScheduler(cfg *config.BotScheduler, repo Repository, channels Channels) *Scheduler {
	return &Scheduler{
		repo:      repo,
		channels:  channels,
		atHours:   cfg.AtHours,
		atMinutes: cfg.AtMinutes,
		atSeconds: cfg.AtSeconds,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	schedule, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	_, err = schedule.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(gocron.NewAtTime(s.atHours, s.atMinutes, s.atSeconds)),
		),
		gocron.NewTask(
			s.SendUpdates,
			ctx,
		),
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
	chats, err := s.repo.GetUpdatesChats(ctx)
	if err != nil {
		slog.Error("failed to get updates chats", slog.Any("error", err))

		return
	}

	for _, chat := range chats {
		updates, err := s.repo.GetAndClearUpdates(ctx, chat)
		if err != nil {
			slog.Error("failed to get updates", slog.Any("error", err))

			return
		}

		msg := tgbotapi.NewMessage(chat, updatesToText(updates))
		msg.ParseMode = tgbotapi.ModeHTML
		msg.DisableWebPagePreview = true
		s.channels.TelegramResp() <- msg
	}
}

func updatesToText(updates []*domain.Update) string {
	builder := strings.Builder{}
	builder.WriteString("Обновления по вашим ссылкам:\n\n")

	for i, update := range updates {
		builder.WriteString(fmt.Sprintf("%d. %s", i+1, update.Message))

		if len(update.Tags) != 0 {
			builder.WriteString(fmt.Sprintf("#%s\n", strings.Join(update.Tags, " #")))
		}

		builder.WriteString("\n")
	}

	return builder.String()
}
