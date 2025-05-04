package updater_test

import (
	"context"
	"errors"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/updater"
	"github.com/es-debug/backend-academy-2024-go-template/internal/application/updater/mocks"
	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestUpdater_UpdatePost(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testUpdate := &domain.Update{}
	errUnavailable := client.NewErrServiceUnavailable(assert.AnError)

	testCases := []struct {
		name          string
		handlers      []updater.Handler
		setupMocks    func(handlers []*mocks.MockHandler)
		expectedError error
	}{
		{
			name: "Success - First handler succeeds",
			handlers: func() []updater.Handler {
				return []updater.Handler{mocks.NewMockHandler(t)}
			}(),
			setupMocks: func(handlers []*mocks.MockHandler) {
				handlers[0].On("UpdatesPost", ctx, testUpdate).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Success - First handler unavailable, second succeeds",
			handlers: func() []updater.Handler {
				return []updater.Handler{mocks.NewMockHandler(t), mocks.NewMockHandler(t)}
			}(),
			setupMocks: func(handlers []*mocks.MockHandler) {
				handlers[0].On("UpdatesPost", ctx, testUpdate).Return(errUnavailable).Once()
				handlers[1].On("UpdatesPost", ctx, testUpdate).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Failure - First handler returns generic error",
			handlers: func() []updater.Handler {
				return []updater.Handler{mocks.NewMockHandler(t), mocks.NewMockHandler(t)}
			}(),
			setupMocks: func(handlers []*mocks.MockHandler) {
				handlers[0].On("UpdatesPost", ctx, testUpdate).Return(assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
		{
			name: "Failure - All handlers unavailable",
			handlers: func() []updater.Handler {
				return []updater.Handler{mocks.NewMockHandler(t), mocks.NewMockHandler(t)}
			}(),
			setupMocks: func(handlers []*mocks.MockHandler) {
				handlers[0].On("UpdatesPost", ctx, testUpdate).Return(errUnavailable).Once()
				handlers[1].On("UpdatesPost", ctx, testUpdate).Return(errUnavailable).Once()
			},
			expectedError: updater.NewErrSendUpdate(),
		},
		{
			name: "Failure - No handlers provided",
			handlers: func() []updater.Handler {
				return []updater.Handler{}
			}(),
			setupMocks:    func(_ []*mocks.MockHandler) {},
			expectedError: updater.NewErrSendUpdate(),
		},
		{
			name: "Failure - First unavailable, second returns generic error",
			handlers: func() []updater.Handler {
				return []updater.Handler{mocks.NewMockHandler(t), mocks.NewMockHandler(t)}
			}(),
			setupMocks: func(handlers []*mocks.MockHandler) {
				handlers[0].On("UpdatesPost", ctx, testUpdate).Return(errUnavailable).Once()
				handlers[1].On("UpdatesPost", ctx, testUpdate).Return(assert.AnError).Once()
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockHandlers := make([]*mocks.MockHandler, len(tc.handlers))
			for i, h := range tc.handlers {
				mockHandlers[i] = h.(*mocks.MockHandler)
			}

			tc.setupMocks(mockHandlers)
			u := updater.New(tc.handlers...)

			err := u.UpdatesPost(ctx, testUpdate)

			if tc.expectedError != nil {
				assert.Error(t, err, "expected error but got nil")
				assert.True(
					t,
					errors.Is(err, tc.expectedError),
					"Expected error %v, got %v",
					tc.expectedError,
					err,
				)

				if !errors.Is(
					err,
					updater.NewErrSendUpdate(),
				) {
					assert.EqualError(
						t,
						err,
						tc.expectedError.Error(),
						"error messages do not match",
					)
				}
			} else {
				assert.NoError(t, err, "expected no error but got %v", err)
			}
		})
	}
}
