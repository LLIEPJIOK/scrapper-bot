package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/application/client/http/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client/kafka"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/mws"
	scrshed "github.com/es-debug/backend-academy-2024-go-template/internal/application/scheduler/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/server/http/health"
	scrapsrv "github.com/es-debug/backend-academy-2024-go-template/internal/application/server/http/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/updater"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/client/github"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/client/sof"
	botapi "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/client"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/producer"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware"
	metricsmw "github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/metrics"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/ratelimiter"
	raterepository "github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/ratelimiter/repository"
)

const local = "local"

type runService = func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup)

func (a *App) services() []runService {
	return []runService{
		a.runServer,
		a.runScheduler,
		a.runCoreKafkaProducer,
		a.runHealthServer,
	}
}

func (a *App) runServer(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("service stopped")

	scrapperServer := scrapsrv.NewServer(a.repo)

	srv, err := scrapperapi.NewServer(scrapperServer)
	if err != nil {
		slog.Error("failed to create scrapper server", slog.Any("error", err))

		return
	}

	repo := raterepository.NewRedis(a.rdb)
	rateLimiter := ratelimiter.NewSlidingWindow(repo, &a.cfg.Scrapper.RateLimiter)

	metricsMW := metricsmw.New(a.Prometheus)
	activeLinksMW := mws.NewLinksCounter(a.Prometheus)

	httpServer := &http.Server{
		Addr:              a.cfg.Scrapper.URL,
		Handler:           middleware.Wrap(srv, metricsMW, rateLimiter, activeLinksMW),
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

	upd, err := a.getBotClient()
	if err != nil {
		slog.Error("failed to create bot client", slog.Any("error", err))

		return
	}

	httpClient := client.New(&a.cfg.Client)
	ghClient := github.New(&a.cfg.GitHub, httpClient)
	sofClient := sof.New(&a.cfg.SOF, httpClient)

	schedule := scrshed.NewScheduler(
		&a.cfg.Scrapper.Scheduler,
		a.repo,
		upd,
		ghClient,
		sofClient,
	)

	if err := schedule.Run(ctx); err != nil {
		slog.Error("failed to run scheduler", slog.Any("error", err))
	}
}

func (a *App) runCoreKafkaProducer(
	ctx context.Context,
	stop context.CancelFunc,
	wg *sync.WaitGroup,
) {
	if a.cfg.App.Env == local {
		return
	}

	defer wg.Done()
	defer stop()
	defer slog.Info("core kafka producer stopped")

	kafkaProducer, err := producer.New(&a.cfg.Kafka.Core, a.channels)
	if err != nil {
		slog.Error("failed to create core kafka producer", slog.Any("error", err))
	}

	if err := kafkaProducer.Run(ctx); err != nil {
		slog.Error("failed to run core kafka producer", slog.Any("error", err))
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
		Addr:              a.cfg.Scrapper.HealthURL,
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

const UpdateHandlerCount = 2

func (a *App) getBotClient() (*updater.Updater, error) {
	handlers := make([]updater.Handler, 0, len(a.cfg.Scrapper.Scheduler.Transports))

	for _, transport := range a.cfg.Scrapper.Scheduler.Transports {
		switch transport {
		case "http":
			httpClient := client.New(&a.cfg.Client)

			ogenClient, err := botapi.NewClient(
				a.cfg.Scrapper.BotURL,
				botapi.WithClient(httpClient),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create http bot client: %w", err)
			}

			handlers = append(handlers, botclient.NewClient(ogenClient))

		case "kafka":
			handlers = append(handlers, kafka.NewProducer(&a.cfg.Kafka, a.channels))
		}
	}

	return updater.New(handlers...), nil
}
