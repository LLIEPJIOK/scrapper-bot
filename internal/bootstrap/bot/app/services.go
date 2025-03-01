package app

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/bot/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/scrapper/client"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/api/http/v1/scrapper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type runService = func(ctx context.Context, stop context.CancelFunc, wg *sync.WaitGroup)

func (a *App) services() []runService {
	return []runService{
		a.runBot,
		a.runProcessor,
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

	ogenClient, err := scrapper.NewClient(
		a.cfg.Scrapper.URL,
		scrapper.WithClient(configureClient(a.cfg)),
	)
	if err != nil {
		slog.Error("failed to create ogen scrapper client", slog.Any("error", err))
	}

	scrap := client.New(ogenClient)
	proc := processor.New(scrap, a.channels)

	if err := proc.Run(ctx); err != nil {
		slog.Error("failed to run processor", slog.Any("err", err))
	}
}

func configureClient(cfg *config.Config) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.Client.DialTimeout,
			KeepAlive: cfg.Client.DialKeepAlive,
		}).DialContext,
		MaxIdleConns:          cfg.Client.MaxIdleConns,
		IdleConnTimeout:       cfg.Client.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.Client.TLSHandshakeTimeout,
		ExpectContinueTimeout: cfg.Client.ExpectContinueTimeout,
		ForceAttemptHTTP2:     true,
		TLSNextProto: make(
			map[string]func(authority string, c *tls.Conn) http.RoundTripper,
		),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Client.Timeout,
	}
}
