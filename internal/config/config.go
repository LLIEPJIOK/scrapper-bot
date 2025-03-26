package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	App       App       `envPrefix:"APP_"`
	Bot       Bot       `envPrefix:"BOT_"`
	Scrapper  Scrapper  `envPrefix:"SCRAPPER_"`
	Client    Client    `envPrefix:"CLIENT_"`
	Server    Server    `envPrefix:"SERVER_"`
	GitHub    GitHub    `envPrefix:"GITHUB_"`
	Scheduler Scheduler `envPrefix:"SCHEDULER_"`
}

type App struct {
	TerminateTimeout time.Duration `env:"TERMINATE_TIMEOUT" envDefault:"5s"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT"  envDefault:"2s"`
}

type Bot struct {
	APIToken    string `env:"API_TOKEN,required"`
	URL         string `env:"URL"                   envDefault:"localhost:8081"`
	ScrapperURL string `env:"SCRAPPER_URL,required"`
}

type Scrapper struct {
	URL    string `env:"URL"              envDefault:"localhost:8080"`
	BotURL string `env:"BOT_URL,required"`
}

type Client struct {
	DialTimeout           time.Duration `env:"DIAL_TIMEOUT"            envDefault:"5s"`
	DialKeepAlive         time.Duration `env:"DIAL_KEEP_ALIVE"         envDefault:"30s"`
	MaxIdleConns          int           `env:"MAX_IDLE_CONNS"          envDefault:"100"`
	IdleConnTimeout       time.Duration `env:"IDLE_CONN_TIMEOUT"       envDefault:"90s"`
	TLSHandshakeTimeout   time.Duration `env:"TLS_HANDSHAKE_TIMEOUT"   envDefault:"10s"`
	ExpectContinueTimeout time.Duration `env:"EXPECT_CONTINUE_TIMEOUT" envDefault:"1s"`
	Timeout               time.Duration `env:"TIMEOUT"                 envDefault:"30s"`
}

type Server struct {
	ReadTimeout       time.Duration `env:"READ_TIMEOUT"        envDefault:"10s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" envDefault:"10s"`
}

type GitHub struct {
	Token string `env:"TOKEN,required"`
}

type Scheduler struct {
	Interval  time.Duration `env:"INTERVAL"   envDefault:"1h"`
	AtHours   uint          `env:"AT_HOURS"   envDefault:"10"`
	AtMinutes uint          `env:"AT_MINUTES" envDefault:"0"`
	AtSeconds uint          `env:"AT_SECONDS" envDefault:"0"`
}

func Load() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	return config, nil
}
