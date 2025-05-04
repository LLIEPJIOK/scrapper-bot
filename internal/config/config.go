package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	clientcfg "github.com/es-debug/backend-academy-2024-go-template/pkg/client/config"
	kafkacfg "github.com/es-debug/backend-academy-2024-go-template/pkg/kafka/config"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/middleware/ratelimiter"
)

type Config struct {
	App      App              `envPrefix:"APP_"`
	Bot      Bot              `envPrefix:"BOT_"`
	Scrapper Scrapper         `envPrefix:"SCRAPPER_"`
	Client   clientcfg.Config `envPrefix:"CLIENT_"`
	Server   Server           `envPrefix:"SERVER_"`
	GitHub   GitHub           `envPrefix:"GITHUB_"`
	SOF      SOF              `envPrefix:"SOF_"`
	Kafka    Kafka            `envPrefix:"KAFKA_"`
}

type App struct {
	Env              string        `env:"ENV"               envDefault:"local"`
	TerminateTimeout time.Duration `env:"TERMINATE_TIMEOUT" envDefault:"5s"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT"  envDefault:"2s"`
}

type Bot struct {
	APIToken    string             `env:"API_TOKEN,required"`
	URL         string             `env:"URL"                   envDefault:":8081"`
	ScrapperURL string             `env:"SCRAPPER_URL,required"`
	Database    Database           `                                               envPrefix:"DATABASE_"`
	Scheduler   BotScheduler       `                                               envPrefix:"SCHEDULER_"`
	Redis       Redis              `                                               envPrefix:"REDIS_"`
	RateLimiter ratelimiter.Config `                                               envPrefix:"RATE_LIMITER_"`
}

type Scrapper struct {
	URL         string             `env:"URL"              envDefault:":8080"`
	BotURL      string             `env:"BOT_URL,required"`
	Database    Database           `                                          envPrefix:"DATABASE_"`
	Scheduler   ScrapperScheduler  `                                          envPrefix:"SCHEDULER_"`
	Redis       Redis              `                                          envPrefix:"REDIS_"`
	RateLimiter ratelimiter.Config `                                          envPrefix:"RATE_LIMITER_"`
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
	Interval  time.Duration `env:"INTERVAL"  envDefault:"1h"`
	PageSize  uint          `env:"PAGE_SIZE" envDefault:"100"`
	Transport string        `env:"TRANSPORT" envDefault:"http"`
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

type Kafka struct {
	Core           kafkacfg.Kafka
	CircuitBreaker CircuitBreaker `envPrefix:"CIRCUIT_BREAKER_"`
	UpdateTopic    string         `                             env:"UPDATE_TOPIC,required"`
}

type Redis struct {
	Address  string `env:"ADDRESS,required"`
	Password string `env:"PASSWORD"`

	DB int `env:"DB" envDefault:"0"`

	DialTimeout  time.Duration `env:"DIAL_TIMEOUT"  envDefault:"5s"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"  envDefault:"3s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"3s"`

	PoolSize     int           `env:"POOL_SIZE"      envDefault:"10"`
	MinIdleConns int           `env:"MIN_IDLE_CONNS" envDefault:"3"`
	PoolTimeout  time.Duration `env:"POOL_TIMEOUT"   envDefault:"4s"`

	IdleTimeout        time.Duration `env:"IDLE_TIMEOUT"         envDefault:"5m"`
	IdleCheckFrequency time.Duration `env:"IDLE_CHECK_FREQUENCY" envDefault:"1m"`

	MaxRetries      int           `env:"MAX_RETRIES"       envDefault:"2"`
	MinRetryBackoff time.Duration `env:"MIN_RETRY_BACKOFF" envDefault:"100ms"`
	MaxRetryBackoff time.Duration `env:"MAX_RETRY_BACKOFF" envDefault:"1s"`

	DefaultTTL time.Duration `env:"DEFAULT_TTL" envDefault:"5m"`
}

type CircuitBreaker struct {
	MaxHalfOpenRequests uint32        `env:"MAX_HALF_OPEN_REQUESTS" envDefault:"5"`
	Interval            time.Duration `env:"INTERVAL"               envDefault:"60s"`
	Timeout             time.Duration `env:"TIMEOUT"                envDefault:"30s"`
	MinRequests         uint32        `env:"MIN_REQUESTS"           envDefault:"10"`
	ConsecutiveFailures uint32        `env:"CONSECUTIVE_FAILURES"   envDefault:"5"`
	FailureRate         float64       `env:"FAILURE_RATE"           envDefault:"0.6"`
}

func Load() (*Config, error) {
	config := &Config{}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	return config, nil
}
