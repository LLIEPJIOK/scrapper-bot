package bot_test

import (
	"errors"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scheduler/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scheduler/bot/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	repository "github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/repository/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_SendUpdates_Error(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)
	repoErr := errors.New("get updates error")
	repoMock.On("GetUpdates").Return(nil, repoErr).Once()

	channels := domain.NewChannels()
	cfg := &config.Scheduler{}

	scheduler := bot.NewScheduler(cfg, repoMock, channels)

	go func() {
		scheduler.SendUpdates()
	}()

	select {
	case msg := <-channels.TelegramResp():
		t.Errorf("Expected no message to be sent, but got: %v", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestScheduler_SendUpdates_Success(t *testing.T) {
	repoMock := mocks.NewMockRepository(t)

	update1 := &repository.UpdateChat{
		ID:    111,
		Links: []string{"link1", "link2"},
	}
	update2 := &repository.UpdateChat{
		ID:    222,
		Links: []string{"link3"},
	}
	updates := []*repository.UpdateChat{update1, update2}
	repoMock.On("GetUpdates").Return(updates, nil).Once()

	channels := domain.NewChannels()
	cfg := &config.Scheduler{}

	scheduler := bot.NewScheduler(cfg, repoMock, channels)

	go func() {
		scheduler.SendUpdates()
	}()

	expectedText1 := "Обновления по вашим ссылкам:\n1. link1\n2. link2\n"
	expectedText2 := "Обновления по вашим ссылкам:\n1. link3\n"

	msg1 := <-channels.TelegramResp()
	m1, ok := msg1.(tgbotapi.MessageConfig)
	require.True(t, ok, "Expected message to be of type MessageConfig")
	assert.Equal(t, update1.ID, m1.ChatID, "Expected ChatID to match for update1")
	assert.Equal(t, expectedText1, m1.Text, "Expected message text to match for update1")

	msg2 := <-channels.TelegramResp()
	m2, ok := msg2.(tgbotapi.MessageConfig)
	require.True(t, ok, "Expected message to be of type MessageConfig")
	assert.Equal(t, update2.ID, m2.ChatID, "Expected ChatID to match for update2")
	assert.Equal(t, expectedText2, m2.Text, "Expected message text to match for update2")

	select {
	case extra := <-channels.TelegramResp():
		t.Errorf("Expected only two messages, but got an extra message: %v", extra)
	case <-time.After(50 * time.Millisecond):
	}
}
