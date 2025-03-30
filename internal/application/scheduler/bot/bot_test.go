package bot_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scheduler/bot"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/scheduler/bot/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScheduler_SendUpdates_GetUpdatesChatsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repoMock := mocks.NewMockRepository(t)
	repoErr := errors.New("get updates error")
	repoMock.On("GetUpdatesChats", ctx, mock.Anything, mock.Anything).Return(nil, repoErr).Once()

	channels := domain.NewChannels()
	cfg := &config.BotScheduler{}

	scheduler := bot.NewScheduler(cfg, repoMock, channels)

	go func() {
		scheduler.SendUpdates(ctx)
	}()

	select {
	case msg := <-channels.TelegramResp():
		t.Errorf("Expected no message to be sent, but got: %v", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestScheduler_SendUpdates_GetUpdatesError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	repoMock := mocks.NewMockRepository(t)
	repoErr := errors.New("get updates error")

	repoMock.On("GetUpdatesChats", ctx, mock.Anything, mock.Anything).Return([]int64{1}, nil).Once()
	repoMock.On("GetUpdates", ctx, int64(1), mock.Anything, mock.Anything).
		Return(nil, repoErr).
		Once()

	channels := domain.NewChannels()
	cfg := &config.BotScheduler{}

	scheduler := bot.NewScheduler(cfg, repoMock, channels)

	go func() {
		scheduler.SendUpdates(ctx)
	}()

	select {
	case msg := <-channels.TelegramResp():
		t.Errorf("Expected no message to be sent, but got: %v", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestScheduler_SendUpdates_Success(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	repoMock := mocks.NewMockRepository(t)

	update1 := domain.Update{
		ChatID:  1,
		URL:     "link1",
		Message: "message1\n",
		Tags:    []string{"tag1"},
	}
	update2 := domain.Update{
		ChatID:  1,
		URL:     "link2",
		Message: "message2\n",
		Tags:    []string{"tag2"},
	}
	update3 := domain.Update{
		ChatID:  2,
		URL:     "link3",
		Message: "message3\n",
	}

	repoMock.On("GetUpdatesChats", ctx, mock.Anything, mock.Anything).
		Return([]int64{1, 2}, nil).
		Once()
	repoMock.On("GetUpdates", ctx, int64(1), mock.Anything, mock.Anything).
		Return([]domain.Update{update1, update2}, nil).
		Once()
	repoMock.On("GetUpdates", ctx, int64(2), mock.Anything, mock.Anything).
		Return([]domain.Update{update3}, nil).
		Once()

	tm := time.Now().Add(time.Second).UTC()
	channels := domain.NewChannels()
	//nolint:gosec //can't overflow
	cfg := &config.BotScheduler{
		AtHours:   uint(tm.Hour()),
		AtMinutes: uint(tm.Minute()),
		AtSeconds: uint(tm.Second()),
	}

	scheduler := bot.NewScheduler(cfg, repoMock, channels)

	go func() {
		expectedText1 := "Обновления по вашим ссылкам:\n\n1. message1\n#tag1\n\n2. message2\n#tag2\n\n"
		expectedText2 := "Обновления по вашим ссылкам:\n\n1. message3\n\n"

		msg1 := <-channels.TelegramResp()
		m1, ok := msg1.(tgbotapi.MessageConfig)
		require.True(t, ok, "Expected message to be of type MessageConfig")
		assert.Equal(t, update1.ChatID, m1.ChatID, "Expected ChatID to match for update1")
		assert.Equal(t, expectedText1, m1.Text, "Expected message text to match for update1")

		msg2 := <-channels.TelegramResp()
		m2, ok := msg2.(tgbotapi.MessageConfig)
		require.True(t, ok, "Expected message to be of type MessageConfig")
		assert.Equal(t, update3.ChatID, m2.ChatID, "Expected ChatID to match for update2")
		assert.Equal(t, expectedText2, m2.Text, "Expected message text to match for update2")

		cancel()
	}()

	err := scheduler.Run(ctx)
	require.NoError(t, err, "Expected Run to not return an error")
}
