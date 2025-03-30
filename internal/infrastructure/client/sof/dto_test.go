package sof_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/client/sof"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnswersToMessages(t *testing.T) {
	t.Parallel()

	questionTitle := "How to <b>test</b> my code?"
	questionLink := "https://stackoverflow.com/q/123456"
	answers := []sof.Answer{
		{
			Owner:     sof.User{DisplayName: "Alice"},
			CreatedAt: time.Date(2025, 3, 30, 15, 30, 0, 0, time.Local).Unix(),
			Body:      strings.Repeat("a", 250),
		},
		{
			Owner:     sof.User{DisplayName: "Bob"},
			CreatedAt: time.Date(2026, 3, 30, 15, 30, 0, 0, time.Local).Unix(),
			Body:      "This is the answer body with <i>HTML</i> content.",
		},
	}

	msgs := sof.AnswersToMessages(answers, questionTitle, questionLink)
	require.Len(t, msgs, 2)

	exp1 := "<b>Новый ответ на StackOverflow!</b>\n" +
		"<b>Вопрос</b>: <a href=\"" + questionLink + "\">How to test my code?</a>\n" +
		"<b>Автор</b>: <i>Alice</i>\n" +
		"<b>Время создания</b>: <i>15:30 30.03.2025</i>\n" +
		fmt.Sprintf("<blockquote>%s</blockquote>\n", strings.Repeat("a", 200)+"...")
	assert.Equal(t, exp1, msgs[0], "invalid message1")

	exp2 := "<b>Новый ответ на StackOverflow!</b>\n" +
		"<b>Вопрос</b>: <a href=\"" + questionLink + "\">How to test my code?</a>\n" +
		"<b>Автор</b>: <i>Bob</i>\n" +
		"<b>Время создания</b>: <i>15:30 30.03.2026</i>\n" +
		"<blockquote>This is the answer body with HTML content.</blockquote>\n"
	assert.Equal(t, exp2, msgs[1], "invalid message2")
}

func TestCommentsToMessages(t *testing.T) {
	t.Parallel()

	questionTitle := "What is <i>Go</i> language?"
	questionLink := "https://stackoverflow.com/q/654321"
	comment := sof.Comment{
		Owner:     sof.User{DisplayName: "Bob"},
		CreatedAt: time.Date(2025, 3, 30, 15, 30, 0, 0, time.Local).Unix(),
		Body:      "This is a comment with <script>alert('x');</script> tags.\n\n",
	}
	comments := []sof.Comment{comment}

	msgs := sof.CommentsToMessages(comments, questionTitle, questionLink)
	require.Len(t, msgs, 1)

	exp := "<b>Новый комментарий на StackOverflow!</b>\n" +
		"<b>Вопрос</b>: <a href=\"" + questionLink + "\">What is Go language?</a>\n" +
		"<b>Автор</b>: <i>Bob</i>\n" +
		"<b>Время создания</b>: <i>15:30 30.03.2025</i>\n" +
		"<blockquote>This is a comment with alert('x'); tags.</blockquote>\n"
	assert.Equal(t, exp, msgs[0], "invalid message")
}
