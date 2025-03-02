package app

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	botclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/bot/client"
	ghclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/github/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/scrapper/scheduler"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/scrapper/service"
	sofclient "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/stackoverflow/client"
	botapi "github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/bot"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
)

type runService = func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup)

func (a *App) services() []runService {
	return []runService{
		a.runServer,
		a.runScheduler,
	}
}

func (a *App) runServer(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	defer stop()
	defer slog.Info("service stopped")

	svc := service.New(a.repo)

	srv, err := scrapper.NewServer(svc)
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
		if err := httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
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

	httpClient := configureClient(&a.cfg.Client)

	ogenClient, err := botapi.NewClient(
		a.cfg.Scrapper.BotURL,
		botapi.WithClient(httpClient),
	)
	if err != nil {
		slog.Error("failed to create ogen bot client", slog.Any("error", err))
	}

	botClient := botclient.New(ogenClient)
	ghClient := ghclient.New(&a.cfg.GitHub, httpClient)
	sofClient := sofclient.New(httpClient)

	schedule := scheduler.New(&a.cfg.Scheduler, a.repo, botClient, ghClient, sofClient)

	if err := schedule.Run(ctx); err != nil {
		slog.Error("failed to run scheduler", slog.Any("error", err))

		return
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
