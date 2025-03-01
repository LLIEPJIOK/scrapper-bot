package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App    App    `yaml:"app"`
	Bot    Bot    `yaml:"bot"`
	Scrapper Scrapper `yaml:"scrapper"`
	Client Client `yaml:"client"`
}

type App struct {
	TerminateTimeout time.Duration `yaml:"terminate_timeout"`
}

type Bot struct {
	APIToken    string
}

type Scrapper struct {
	URL string `yaml:"url"`
}

type Client struct {
	DialTimeout           time.Duration `yaml:"dial_timeout"`
	DialKeepAlive         time.Duration `yaml:"dial_keep_alive"`
	MaxIdleConns          int           `yaml:"max_idle_conns"`
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   time.Duration `yaml:"tls_handshake_timeout"`
	ExpectContinueTimeout time.Duration `yaml:"expect_continue_timeout"`
	Timeout               time.Duration `yaml:"timeout"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, NewErrConfig("config path not specified")
	}

	path += "/default.yaml"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, NewErrConfig("config file not found")
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, NewErrConfig(fmt.Sprintf("failed to open config file: %s", err))
	}

	defer func() {
		if err := f.Close(); err != nil {
			slog.Error(
				"failed to close config file",
				slog.String("path", path),
				slog.Any("err", err),
			)
		}
	}()

	config := &Config{}

	err = yaml.NewDecoder(f).Decode(config)
	if err != nil {
		return nil, NewErrConfig(fmt.Sprintf("failed to decode config: %s", err))
	}

	config.Bot.APIToken = os.Getenv("BOT_API_TOKEN")

	return config, nil
}
