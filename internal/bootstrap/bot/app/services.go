package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/http/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/kafka"
	botscheduler "github.com/es-debug/backend-academy-2024-go-template/internal/application/scheduler/bot"
	botsrv "github.com/es-debug/backend-academy-2024-go-template/internal/application/server/http/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/server/http/health"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	botapi "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/client"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/consumer"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware"
	metricsmw "github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/metrics"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/ratelimiter"
	raterepository "github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/ratelimiter/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const local = "local"

type runService = func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup)

func (a *App) services() []runService {
	return []runService{
		a.runBot,
		a.runProcessor,
		a.runServer,
		a.runScheduler,
		a.runCoreKafkaConsumer,
		a.runAppKafkaConsumer,
		a.runHealthServer,
	}
}

func (a *App) runBot(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("bot stopped")

	api, err := tgbotapi.NewBotAPI(a.cfg.Bot.APIToken)
	if err != nil {
		slog.Error("failed to create bot api", slog.Any("err", err))

		return
	}

	api.Debug = true

	slog.Info("bot started", slog.Any("username", api.Self.UserName))

	tgBot, err := bot.New(api, a.channels)
	if err != nil {
		slog.Error("failed to create bot", slog.Any("err", err))

		return
	}

	if err := tgBot.Run(ctx); err != nil {
		slog.Error("failed to run bot", slog.Any("err", err))
	}
}

func (a *App) runProcessor(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("processor stopped")

	ogenClient, err := scrapperapi.NewClient(
		a.cfg.Bot.ScrapperURL,
		scrapperapi.WithClient(client.New(&a.cfg.Client)),
	)
	if err != nil {
		slog.Error("failed to create ogen scrapper client", slog.Any("error", err))
	}

	scrap := scrapper.NewClient(ogenClient)
	proc := processor.New(scrap, a.channels, a.cache, a.prometheus)

	if err := proc.Run(ctx); err != nil {
		slog.Error("failed to run processor", slog.Any("err", err))
	}
}

func (a *App) runServer(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("service stopped")

	botServer := botsrv.NewServer(a.cache, a.channels)

	srv, err := botapi.NewServer(botServer)
	if err != nil {
		slog.Error("failed to create bot server", slog.Any("error", err))

		return
	}

	repo := raterepository.NewRedis(a.rdb)
	rateLimiter := ratelimiter.NewSlidingWindow(repo, &a.cfg.Bot.RateLimiter)

	metrics := metricsmw.New(a.prometheus)

	httpServer := &http.Server{
		Addr:              a.cfg.Bot.URL,
		Handler:           middleware.Wrap(srv, metrics, rateLimiter),
		ReadTimeout:       a.cfg.Server.ReadTimeout,
		ReadHeaderTimeout: a.cfg.Server.ReadHeaderTimeout,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start scrapper server", slog.Any("error", err))

			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.App.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown scrapper server", slog.Any("error", err))
	}
}

func (a *App) runScheduler(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("scheduler stopped")

	schedule := botscheduler.NewScheduler(&a.cfg.Bot.Scheduler, a.cache, a.channels)

	if err := schedule.Run(ctx); err != nil {
		slog.Error("failed to run scheduler", slog.Any("error", err))

		return
	}
}

func (a *App) runCoreKafkaConsumer(
	ctx context.Context,
	stop context.CancelFunc,
	wg *sync.WaitGroup,
) {
	if a.cfg.App.Env == local {
		return
	}

	defer wg.Done()
	defer stop()
	defer slog.Info("core kafka consumer stopped")

	core, err := consumer.New(&a.cfg.Kafka.Core, a.db, a.channels)
	if err != nil {
		slog.Error("failed to create core kafka consumer", slog.Any("error", err))

		return
	}

	if err := core.Run(ctx); err != nil {
		slog.Error("failed to run core kafka consumer", slog.Any("error", err))
	}
}

func (a *App) runAppKafkaConsumer(
	ctx context.Context,
	stop context.CancelFunc,
	wg *sync.WaitGroup,
) {
	if a.cfg.App.Env == local {
		return
	}

	defer wg.Done()
	defer stop()
	defer slog.Info("app kafka consumer stopped")

	kafkaConsumer := kafka.NewConsumer(a.cache, a.channels)

	if err := kafkaConsumer.Run(ctx); err != nil {
		slog.Error("failed to run app kafka consumer", slog.Any("error", err))
	}
}

func (a *App) runHealthServer(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("health server stopped")

	srv := http.NewServeMux()
	ctrl := health.New()
	ctrl.RegisterRoutes(srv)

	httpServer := &http.Server{
		Addr:              a.cfg.Bot.HealthURL,
		Handler:           srv,
		ReadTimeout:       a.cfg.Server.ReadTimeout,
		ReadHeaderTimeout: a.cfg.Server.ReadHeaderTimeout,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("failed to start health server", slog.Any("error", err))

			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.App.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown health server", slog.Any("error", err))
	}
}
