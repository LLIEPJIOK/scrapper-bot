package main

import (
	"context"
	"flag"
	"os"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bootstrap/bot/app"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"golang.org/x/exp/slog"
)

const (
	OkCode = iota
	ErrorConfigLoad
	ErrorCreateApp
	ErrorRunApp
)

func main() {
	var configPath string

	flag.StringVar(&configPath, "config", "./config", "Path to JSON config file")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("Error loading config", slog.Any("error", err))
		os.Exit(ErrorConfigLoad)
	}

	application := app.New(cfg)

	if runerr := application.Run(ctx); runerr != nil {
		slog.Error("Error running application", slog.Any("error", runerr))

		os.Exit(ErrorRunApp)
	}

	os.Exit(OkCode)
}
