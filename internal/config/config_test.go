package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_Success(t *testing.T) {
	assert.NoError(t, os.Setenv("BOT_API_TOKEN", "test_token"))
	assert.NoError(t, os.Setenv("BOT_SCRAPPER_URL", "http://localhost:8080"))
	assert.NoError(t, os.Setenv("SCRAPPER_BOT_URL", "http://localhost:8081"))
	assert.NoError(t, os.Setenv("GITHUB_TOKEN", "github_test_token"))

	assert.NoError(t, os.Setenv("APP_TERMINATE_TIMEOUT", "7s"))
	assert.NoError(t, os.Setenv("APP_SHUTDOWN_TIMEOUT", "3s"))
	assert.NoError(t, os.Setenv("CLIENT_DIAL_TIMEOUT", "6s"))
	assert.NoError(t, os.Setenv("SERVER_READ_TIMEOUT", "12s"))
	assert.NoError(t, os.Setenv("SCRAPPER_SCHEDULER_INTERVAL", "2h"))

	config, err := config.Load()
	assert.NoError(t, err, "expected no error loading configuration")

	assert.Equal(t, 7*time.Second, config.App.TerminateTimeout, "unexpected App.TerminateTimeout")
	assert.Equal(t, 3*time.Second, config.App.ShutdownTimeout, "unexpected App.ShutdownTimeout")

	assert.Equal(t, "test_token", config.Bot.APIToken, "unexpected Bot.APIToken")
	assert.Equal(t, "localhost:8081", config.Bot.URL, "unexpected default Bot.URL")
	assert.Equal(t, "http://localhost:8080", config.Bot.ScrapperURL, "unexpected Bot.ScrapperURL")

	assert.Equal(t, "localhost:8080", config.Scrapper.URL, "unexpected default Scrapper.URL")
	assert.Equal(t, "http://localhost:8081", config.Scrapper.BotURL, "unexpected Scrapper.BotURL")

	assert.Equal(t, 6*time.Second, config.Client.DialTimeout, "unexpected Client.DialTimeout")

	assert.Equal(t, 12*time.Second, config.Server.ReadTimeout, "unexpected Server.ReadTimeout")
	assert.Equal(
		t,
		10*time.Second,
		config.Server.ReadHeaderTimeout,
		"unexpected default Server.ReadHeaderTimeout",
	)

	assert.Equal(t, "github_test_token", config.GitHub.Token, "unexpected GitHub.Token")

	assert.Equal(
		t,
		2*time.Hour,
		config.Scrapper.Scheduler.Interval,
		"unexpected Scheduler.Interval",
	)
	assert.Equal(t, uint(10), config.Bot.Scheduler.AtHours, "unexpected default Bot.Scheduler.AtHours")
	assert.Equal(
		t,
		uint(0),
		config.Bot.Scheduler.AtMinutes,
		"unexpected default Bot.Scheduler.AtMinutes",
	)
	assert.Equal(
		t,
		uint(0),
		config.Bot.Scheduler.AtSeconds,
		"unexpected default Bot.Scheduler.AtSeconds",
	)
}

func TestLoad_MissingRequired(t *testing.T) {
	os.Unsetenv("BOT_API_TOKEN")
	os.Unsetenv("BOT_SCRAPPER_URL")
	os.Unsetenv("SCRAPPER_BOT_URL")
	os.Unsetenv("GITHUB_TOKEN")

	_, err := config.Load()
	assert.Error(t, err, "expected error due to missing required environment variables")
}
