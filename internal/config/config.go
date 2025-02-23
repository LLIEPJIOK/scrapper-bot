package config

import (
	"fmt"
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Bot    Bot    `yaml:"bot"`
	Client Client `yaml:"client"`
}

type Bot struct {
	APIToken    string
	ScrapperURL string `yaml:"scrapper_url"`
}

type Client struct {
	DialTimeoutSeconds           int `yaml:"dial_timeout_seconds"`
	DialKeepAliveSeconds         int `yaml:"dial_keep_alive_seconds"`
	MaxIdleConns                 int `yaml:"max_idle_conns"`
	IdleConnTimeoutSeconds       int `yaml:"idle_conn_timeout_seconds"`
	TLSHandshakeTimeoutSeconds   int `yaml:"tls_handshake_timeout_seconds"`
	ExpectContinueTimeoutSeconds int `yaml:"expect_continue_timeout_seconds"`
	Timeout                      int `yaml:"timeout"`
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
