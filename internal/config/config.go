package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	App      App      `envPrefix:"APP_"`
	Bot      Bot      `envPrefix:"BOT_"`
	Scrapper Scrapper `envPrefix:"SCRAPPER_"`
	Client   Client   `envPrefix:"CLIENT_"`
	Server   Server   `envPrefix:"SERVER_"`
	GitHub   GitHub   `envPrefix:"GITHUB_"`
	SOF      SOF      `envPrefix:"SOF_"`
}

type App struct {
	TerminateTimeout time.Duration `env:"TERMINATE_TIMEOUT" envDefault:"5s"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT"  envDefault:"2s"`
}

type Bot struct {
	APIToken    string       `env:"API_TOKEN,required"`
	URL         string       `env:"URL"                   envDefault:":8081"`
	ScrapperURL string       `env:"SCRAPPER_URL,required"`
	Database    Database     `                                               envPrefix:"DATABASE_"`
	Scheduler   BotScheduler `                                               envPrefix:"SCHEDULER_"`
}

type Scrapper struct {
	URL       string            `env:"URL"              envDefault:":8080"`
	BotURL    string            `env:"BOT_URL,required"`
	Database  Database          `                                          envPrefix:"DATABASE_"`
	Scheduler ScrapperScheduler `                                          envPrefix:"SCHEDULER_"`
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
	Token    string `env:"TOKEN,required"`
	PageSize string `env:"PAGE_SIZE"      envDefault:"100"`
}

type SOF struct {
	PageSize string `env:"PAGE_SIZE" envDefault:"100"`
}

type BotScheduler struct {
	AtHours   uint `env:"AT_HOURS"   envDefault:"10"`
	AtMinutes uint `env:"AT_MINUTES" envDefault:"0"`
	AtSeconds uint `env:"AT_SECONDS" envDefault:"0"`
}

type ScrapperScheduler struct {
	Interval time.Duration `env:"INTERVAL"  envDefault:"1h"`
	PageSize uint          `env:"PAGE_SIZE" envDefault:"100"`
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

func Load() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	return config, nil
}
