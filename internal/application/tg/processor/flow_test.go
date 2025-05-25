package processor_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTrackFlow(t *testing.T) {
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	cache := mocks.NewMockCache(t)
	metrics := mocks.NewMockMetrics(t)

	metrics.On("IncTGRequestsTotal", "callback", "success").Times(4)
	metrics.On("ObserveTGRequestsDurationSeconds", "callback", mock.Anything).Times(4)

	ctx, cancel := context.WithCancel(context.Background())
	proc := processor.New(client, channels, cache, metrics)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := proc.Run(ctx)
		assert.NoError(t, err, "Run should not return error")
	}()

	// command
	metrics.On("IncTGRequestsTotal", "command", "success").Once()
	metrics.On("ObserveTGRequestsDurationSeconds", "command", mock.Anything).Once()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "/track",
		Type:    domain.Command,
	}

	// track
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	// trackAddLink
	client.On("GetLinks", ctx, int64(1), "").Return(nil, nil)
	metrics.On("IncTGRequestsTotal", "track_add_link", "success").Once()
	metrics.On("ObserveTGRequestsDurationSeconds", "track_add_link", mock.Anything).Once()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "https://github.com/LLIEPJIOK/nginxparser",
		Type:    domain.Message,
	}
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	// trackAddTags
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "track_add_tags",
		Type:    domain.Callback,
	}
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	metrics.On("IncTGRequestsTotal", "track_add_tags", "success").Once()
	metrics.On("ObserveTGRequestsDurationSeconds", "track_add_tags", mock.Anything).Once()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "tag1 tag2",
		Type:    domain.Message,
	}
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	// trackAddFilters
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "track_add_filters",
		Type:    domain.Callback,
	}
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	metrics.On("IncTGRequestsTotal", "track_add_filters", "success").Once()
	metrics.On("ObserveTGRequestsDurationSeconds", "track_add_filters", mock.Anything).Once()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "filter1 filter2",
		Type:    domain.Message,
	}
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	// trackAddSetTime
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "track_add_set_time",
		Type:    domain.Callback,
	}
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "track_add_set_time_immediately",
		Type:    domain.Callback,
	}

	// save
	client.On("AddLink", ctx, mock.MatchedBy(func(link *domain.Link) bool {
		return link.URL == "https://github.com/LLIEPJIOK/nginxparser" && link.ChatID == 1 &&
			link.Filters[0] == "filter1" && link.Filters[1] == "filter2" &&
			link.Tags[0] == "tag1" && link.Tags[1] == "tag2"
	})).Return(nil).Once()
	cache.On("InvalidateListLinks", mock.Anything, int64(1)).
		Return(nil).
		Once()

	<-channels.TelegramResp()

	cancel()
	wg.Wait()
}

func TestUntrackFlow(t *testing.T) {
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	cache := mocks.NewMockCache(t)
	metrics := mocks.NewMockMetrics(t)

	ctx, cancel := context.WithCancel(context.Background())
	proc := processor.New(client, channels, cache, metrics)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		err := proc.Run(ctx)
		assert.NoError(t, err, "Run should not return error")
	}()

	// command
	metrics.On("IncTGRequestsTotal", "command", "success").Once()
	metrics.On("ObserveTGRequestsDurationSeconds", "command", mock.Anything).Once()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "/untrack",
		Type:    domain.Command,
	}

	time.Sleep(100 * time.Millisecond) // wait for processing

	// untrack
	<-channels.TelegramResp()

	time.Sleep(100 * time.Millisecond) // wait for processing

	// trackDeleteLink
	metrics.On("IncTGRequestsTotal", "untrack_delete_link", "success").Once()
	metrics.On("ObserveTGRequestsDurationSeconds", "untrack_delete_link", mock.Anything).Once()
	channels.TelegramReq() <- domain.TelegramRequest{
		ChatID:  1,
		Message: "https://github.com/LLIEPJIOK/nginxparser",
		Type:    domain.Message,
	}
	client.On("DeleteLink", ctx, int64(1), "https://github.com/LLIEPJIOK/nginxparser").
		Return(nil).
		Once()
	cache.On("InvalidateListLinks", mock.Anything, int64(1)).
		Return(nil).
		Once()
	<-channels.TelegramResp()

	cancel()
	wg.Wait()
}
