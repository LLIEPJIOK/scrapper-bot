package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/application/http/client/bot"
	scrapsrv "github.com/es-debug/backend-academy-2024-go-template/internal/application/http/server/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/kafka"
	scrshed "github.com/es-debug/backend-academy-2024-go-template/internal/application/scheduler/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/client/github"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/client/sof"
	botapi "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	scrapperapi "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/producer"
)

const kafkaTransport = "kafka"

type runService = func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup)

func (a *App) services() []runService {
	return []runService{
		a.runServer,
		a.runScheduler,
		a.runCoreKafkaProducer,
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

	httpServer := &http.Server{
		Addr:              a.cfg.Scrapper.URL,
		Handler:           srv,
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

	botClient, err := a.getBotClient()
	if err != nil {
		slog.Error("failed to create bot client", slog.Any("error", err))

		return
	}

	httpClient := configureClient(&a.cfg.Client)
	ghClient := github.New(&a.cfg.GitHub, httpClient)
	sofClient := sof.New(&a.cfg.SOF, httpClient)

	schedule := scrshed.NewScheduler(
		&a.cfg.Scrapper.Scheduler,
		a.repo,
		botClient,
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
	if a.cfg.Scrapper.Scheduler.Transport != kafkaTransport {
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

func configureClient(cfg *config.Client) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: cfg.DialKeepAlive,
		}).DialContext,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,
		ForceAttemptHTTP2:     true,
		TLSNextProto: make(
			map[string]func(authority string, c *tls.Conn) http.RoundTripper,
		),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

func (a *App) getBotClient() (scrshed.Client, error) {
	switch a.cfg.Scrapper.Scheduler.Transport {
	case "http":
		httpClient := configureClient(&a.cfg.Client)

		ogenClient, err := botapi.NewClient(
			a.cfg.Scrapper.BotURL,
			botapi.WithClient(httpClient),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create http bot client: %w", err)
		}

		botClient := botclient.NewClient(ogenClient)

		return botClient, nil
	case "kafka":
		return kafka.NewProducer(a.cfg.Kafka.UpdateTopic, a.channels), nil
	}

	return nil, fmt.Errorf("unknown transport: %s", a.cfg.Scrapper.Scheduler.Transport)
}
