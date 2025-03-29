package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/pressly/goose/v3"

	_ "github.com/lib/pq"
)

const (
	OkCode = iota
	ErrorConfigLoad
	ErrorConnectDatabase
	ErrorMigrate
)

type Config struct {
	BotDB      Database `envPrefix:"BOT_DATABASE_"`
	ScrapperDB Database `envPrefix:"SCRAPPER_DATABASE_"`
}

type Database struct {
	Host     string `env:"HOST,required"`
	Port     string `env:"PORT,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	Name     string `env:"NAME,required"`
	SSLMode  string `env:"SSL_MODE,required"`
	Type     string `env:"TYPE,required"`
}

func main() {
	var cmd string

	flag.StringVar(&cmd, "command", "up", "Migration command")
	flag.Parse()

	var config Config

	if err := env.Parse(&config); err != nil {
		slog.Error("Error loading config", slog.Any("error", err))
		os.Exit(ErrorConfigLoad)
	}

	if code := botMigrate(&config, cmd); code != OkCode {
		os.Exit(code)
	}

	os.Exit(scrapperMigrate(&config, cmd))
}

func botMigrate(cfg *Config, cmd string) (code int) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.BotDB.Host,
		cfg.BotDB.Port,
		cfg.BotDB.User,
		cfg.BotDB.Password,
		cfg.BotDB.Name,
		cfg.BotDB.SSLMode,
	)

	return migrate(dsn, "bot", cmd)
}

func scrapperMigrate(cfg *Config, cmd string) (code int) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.ScrapperDB.Host,
		cfg.ScrapperDB.Port,
		cfg.ScrapperDB.User,
		cfg.ScrapperDB.Password,
		cfg.ScrapperDB.Name,
		cfg.ScrapperDB.SSLMode,
	)

	return migrate(dsn, "scrapper", cmd)
}

func migrate(dsn, tpe, cmd string) (code int) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error(fmt.Sprintf("Error connect to %s database", tpe), slog.Any("error", err))

		return ErrorConnectDatabase
	}

	if err := db.Ping(); err != nil {
		slog.Error(fmt.Sprintf("Error ping %s database", tpe), slog.Any("error", err))

		return ErrorConnectDatabase
	}

	defer func() {
		if err := db.Close(); err != nil {
			slog.Error(fmt.Sprintf("Error close %s database", tpe), slog.Any("error", err))
		}
	}()

	if err = goose.RunContext(context.Background(), cmd, db, "./migrations/"+tpe); err != nil {
		slog.Error(fmt.Sprintf("Error migrate %s database", tpe), slog.Any("error", err))

		return ErrorMigrate
	}

	return OkCode
}
