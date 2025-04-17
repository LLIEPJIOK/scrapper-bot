package processor_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/tg/processor/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleTrackLinkAdder(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name  string
		state *processor.State
		exp   *fsm.Result[*processor.State]
	}{
		{
			name: "sof link",
			state: &processor.State{
				Message: "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
				ChatID:  1,
				Object: &domain.Link{
					ChatID: 1,
				},
			},
			exp: &fsm.Result[*processor.State]{
				NextState:        "callback",
				IsAutoTransition: false,
				Result: &processor.State{
					Message: "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
					ChatID:  1,
					Object: &domain.Link{
						URL:    "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
						ChatID: 1,
					},
				},
			},
		},
		{
			name: "gh repo link",
			state: &processor.State{
				Message: "https://github.com/LLIEPJIOK/LLIEPJIOK",
				ChatID:  1,
				Object: &domain.Link{
					ChatID: 1,
				},
			},
			exp: &fsm.Result[*processor.State]{
				NextState:        "callback",
				IsAutoTransition: false,
				Result: &processor.State{
					Message: "https://github.com/LLIEPJIOK/LLIEPJIOK",
					ChatID:  1,
					Object: &domain.Link{
						URL:    "https://github.com/LLIEPJIOK/LLIEPJIOK",
						ChatID: 1,
					},
				},
			},
		},
		// {
		// 	name: "gh issue link",
		// 	state: &processor.State{
		// 		Message: "https://github.com/LLIEPJIOK/forum/issues/1",
		// 		ChatID:  1,
		// 		Object: &domain.Link{
		// 			ChatID: 1,
		// 		},
		// 	},
		// 	exp: &fsm.Result[*processor.State]{
		// 		NextState:        "callback",
		// 		IsAutoTransition: false,
		// 		Result: &processor.State{
		// 			Message: "https://github.com/LLIEPJIOK/forum/issues/1",
		// 			ChatID:  1,
		// 			Object: &domain.Link{
		// 				URL:    "https://github.com/LLIEPJIOK/forum/issues/1",
		// 				ChatID: 1,
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	name: "gh pr link",
		// 	state: &processor.State{
		// 		Message: "https://github.com/aleksander-git/telegram-torrent/pull/7",
		// 		ChatID:  1,
		// 		Object: &domain.Link{
		// 			ChatID: 1,
		// 		},
		// 	},
		// 	exp: &fsm.Result[*processor.State]{
		// 		NextState:        "callback",
		// 		IsAutoTransition: false,
		// 		Result: &processor.State{
		// 			Message: "https://github.com/aleksander-git/telegram-torrent/pull/7",
		// 			ChatID:  1,
		// 			Object: &domain.Link{
		// 				URL:    "https://github.com/aleksander-git/telegram-torrent/pull/7",
		// 				ChatID: 1,
		// 			},
		// 		},
		// 	},
		// },
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			channels := domain.NewChannels()
			client := mocks.NewMockClient(t)
			client.On("GetLinks", ctx, int64(1), "").Return(nil, nil)

			handler := processor.NewTrackLinkAdder(client, channels)

			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				defer wg.Done()

				ans := <-channels.TelegramResp()
				msg, ok := ans.(tgbotapi.MessageConfig)
				require.True(t, ok, "not tg message")
				assert.Equal(
					t,
					"Можете добавить опциональные поля или сохранить ссылку в текущем состоянии.",
					msg.Text,
					"wrong message",
				)

				keyboard, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
				require.True(t, ok, "not tg keyboard")
				assert.Equal(t, 3, len(keyboard.InlineKeyboard), "wrong keyboard")
			}()

			res := handler.Handle(ctx, tc.state)
			assert.Equal(t, tc.exp, res, "wrong result")

			wg.Wait()
		})
	}
}

func TestHandleTrackLinkAdderInvalidLinks(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name  string
		state *processor.State
		exp   *fsm.Result[*processor.State]
	}{
		{
			name: "sof invalid link",
			state: &processor.State{
				FSMState: "track_add_link",
				Message:  "https://stackoverflow.com/questions/id/androidmanifest-xml-file-raising-errors-with-no-exception",
				ChatID:   1,
				Object: &domain.Link{
					ChatID: 1,
				},
			},
			exp: &fsm.Result[*processor.State]{
				NextState:        "track_add_link",
				IsAutoTransition: false,
				Result: &processor.State{
					FSMState: "track_add_link",
					Message:  "https://stackoverflow.com/questions/id/androidmanifest-xml-file-raising-errors-with-no-exception",
					ChatID:   1,
					Object: &domain.Link{
						ChatID: 1,
					},
				},
			},
		},
		{
			name: "gh invalid link",
			state: &processor.State{
				FSMState: "track_add_link",
				Message:  "https://github.com/aleksander-git/telegram-torrent/pulls",
				ChatID:   1,
				Object: &domain.Link{
					ChatID: 1,
				},
			},
			exp: &fsm.Result[*processor.State]{
				NextState:        "track_add_link",
				IsAutoTransition: false,
				Result: &processor.State{
					FSMState: "track_add_link",
					Message:  "https://github.com/aleksander-git/telegram-torrent/pulls",
					ChatID:   1,
					Object: &domain.Link{
						ChatID: 1,
					},
				},
			},
		},
		{
			name: "invalid link",
			state: &processor.State{
				FSMState: "track_add_link",
				Message:  "invalid",
				ChatID:   1,
				Object: &domain.Link{
					ChatID: 1,
				},
			},
			exp: &fsm.Result[*processor.State]{
				NextState:        "track_add_link",
				IsAutoTransition: false,
				Result: &processor.State{
					FSMState: "track_add_link",
					Message:  "invalid",
					ChatID:   1,
					Object: &domain.Link{
						ChatID: 1,
					},
				},
			},
		},
		{
			name: "another link",
			state: &processor.State{
				FSMState: "track_add_link",
				Message:  "https://docs.github.com/en/rest/branches/branches?apiVersion=2022-11-28",
				ChatID:   1,
				Object: &domain.Link{
					ChatID: 1,
				},
			},
			exp: &fsm.Result[*processor.State]{
				NextState:        "track_add_link",
				IsAutoTransition: false,
				Result: &processor.State{
					FSMState: "track_add_link",
					Message:  "https://docs.github.com/en/rest/branches/branches?apiVersion=2022-11-28",
					ChatID:   1,
					Object: &domain.Link{
						ChatID: 1,
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			channels := domain.NewChannels()
			client := mocks.NewMockClient(t)
			handler := processor.NewTrackLinkAdder(client, channels)

			go func() {
				ans := <-channels.TelegramResp()
				msg, ok := ans.(tgbotapi.MessageConfig)
				require.True(t, ok, "not tg message")

				text := `Неверный формат ссылки. Используйте следующие форматы:
- https://stackoverflow.com/questions/{id}/{title}
- https://github.com/{user}/{repo}`
				assert.Equal(
					t,
					text,
					msg.Text,
				)
			}()

			res := handler.Handle(ctx, tc.state)
			assert.Equal(t, tc.exp, res, "wrong result")
		})
	}
}

func TestHandleInvalidObject(t *testing.T) {
	t.Parallel()

	state := &processor.State{
		Message: "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
		ChatID:  1,
		Object:  "invalid",
	}
	exp := &fsm.Result[*processor.State]{
		NextState:        "fail",
		IsAutoTransition: true,
		Result: &processor.State{
			Message: "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
			ChatID:  1,
			Object:  "invalid",
		},
	}

	ctx := context.Background()
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("GetLinks", ctx, int64(1), "").Return(nil, nil)

	handler := processor.NewTrackLinkAdder(client, channels)

	res := handler.Handle(ctx, state)
	assert.Equal(t, exp, res, "wrong result")
}

func TestTrackLinkAdder_Handle_LinkExists(t *testing.T) {
	t.Parallel()

	state := &processor.State{
		FSMState: "track_add_link",
		Message:  "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
		ChatID:   1,
		Object: &domain.Link{
			ChatID: 1,
		},
	}
	exp := &fsm.Result[*processor.State]{
		NextState:        "track_add_link",
		IsAutoTransition: false,
		Result: &processor.State{
			FSMState: "track_add_link",
			Message:  "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
			ChatID:   1,
			Object: &domain.Link{
				ChatID: 1,
			},
		},
	}

	ctx := context.Background()
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("GetLinks", ctx, int64(1), "").Return([]*domain.Link{
		{
			URL: "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
		},
	}, nil)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		ans := <-channels.TelegramResp()
		msg, ok := ans.(tgbotapi.MessageConfig)
		require.True(t, ok, "not tg message")

		text := `Ссылка уже существует. Введите другую ссылку или посмотрите список, используя /list`
		assert.Equal(
			t,
			text,
			msg.Text,
		)
	}()

	handler := processor.NewTrackLinkAdder(client, channels)

	res := handler.Handle(ctx, state)
	assert.Equal(t, exp, res, "wrong result")
}

func TestTrackLinkAdder_Handle_GetLinksError(t *testing.T) {
	t.Parallel()

	state := &processor.State{
		FSMState: "track_add_link",
		Message:  "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
		ChatID:   1,
		Object: &domain.Link{
			ChatID: 1,
		},
	}
	exp := &fsm.Result[*processor.State]{
		NextState:        "fail",
		IsAutoTransition: true,
		Result: &processor.State{
			FSMState: "track_add_link",
			Message:  "https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception",
			ChatID:   1,
			Object: &domain.Link{
				ChatID: 1,
			},
			ShowError: "ошибка при добавлении ссылки",
		},
		Error: errors.New(
			`h.IsLinkExists(` +
				`ctx, ` +
				`"https://stackoverflow.com/questions/79476948/androidmanifest-xml-file-raising-errors-with-no-exception", ` +
				`1` +
				`): failed to get links: test error`,
		),
	}

	ctx := context.Background()
	channels := domain.NewChannels()
	client := mocks.NewMockClient(t)
	client.On("GetLinks", ctx, int64(1), "").Return(nil, errors.New("test error"))

	wg := sync.WaitGroup{}
	wg.Add(1)

	handler := processor.NewTrackLinkAdder(client, channels)

	res := handler.Handle(ctx, state)
	assert.Equal(t, res.NextState, exp.NextState, "wrong next state")
	assert.Equal(t, res.IsAutoTransition, exp.IsAutoTransition, "wrong is auto transition")
	assert.Equal(t, res.Result, exp.Result, "wrong result")
	assert.EqualError(t, res.Error, exp.Error.Error(), "wrong error")
}
