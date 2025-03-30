package processor_test

import (
	"context"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrackFlow(t *testing.T) {
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	proc := processor.New(client, channels)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := proc.Run(ctx)
		assert.NoError(t, err, "Run should not return error")
	}()

	// command
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "/track",
		Type:    domain.Command,
	}

	// track
	<-channels.TelegramResp()

	// trackAddLink
	client.On("GetLinks", ctx, int64(1), "").Return(nil, nil)
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "https://github.com/LLIEPJIOK/nginxparser",
		Type:    domain.Message,
	}
	<-channels.TelegramResp()

	// trackAddTags
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "track_add_tags",
		Type:    domain.Callback,
	}
	<-channels.TelegramResp()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "tag1 tag2",
		Type:    domain.Message,
	}
	<-channels.TelegramResp()

	// trackAddFilters
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "track_add_filters",
		Type:    domain.Callback,
	}
	<-channels.TelegramResp()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "filter1 filter2",
		Type:    domain.Message,
	}

	// save
	client.On("AddLink", ctx, mock.MatchedBy(func(link *domain.Link) bool {
		return link.URL == "https://github.com/LLIEPJIOK/nginxparser" && link.ChatID == 1 &&
			link.Filters[0] == "filter1" && link.Filters[1] == "filter2" &&
			link.Tags[0] == "tag1" && link.Tags[1] == "tag2"
	})).Return(nil).Once()
	<-channels.TelegramResp()

	cancel()
	wg.Wait()
}

func TestUntrackFlow(t *testing.T) {
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	proc := processor.New(client, channels)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := proc.Run(ctx)
		assert.NoError(t, err, "Run should not return error")
	}()

	// command
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "/untrack",
		Type:    domain.Command,
	}

	// untrack
	<-channels.TelegramResp()

	// trackDeleteLink
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "https://github.com/LLIEPJIOK/nginxparser",
		Type:    domain.Message,
	}
	client.On("DeleteLink", ctx, int64(1), "https://github.com/LLIEPJIOK/nginxparser").
		Return(nil).
		Once()
	<-channels.TelegramResp()

	cancel()
	wg.Wait()
}
