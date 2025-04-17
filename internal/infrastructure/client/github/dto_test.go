package github_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/infrastructure/client/github"
	"github.com/stretchr/testify/assert"
)

func TestDataToMessage_PR(t *testing.T) {
	data := github.Data{
		PR: github.PullRequest{
			URL: "https://github.com/example/repo/pull/1",
		},
		URL:       "https://github.com/example/repo/pull/1",
		Title:     "Add new feature",
		User:      github.User{Login: "devUser"},
		CreatedAt: time.Date(2025, 3, 30, 15, 30, 0, 0, time.Local),
		Body:      "This is the body of the pull request. It contains details about the changes.",
	}

	msg := github.DataToMessage(&data)

	exp := "<b>Новый Pull Request на Github!</b>\n" +
		"<b>Название</b>: <a href=\"https://github.com/example/repo/pull/1\">Add new feature</a>\n" +
		"<b>Автор</b>: <i>devUser</i>\n" +
		"<b>Время создания</b>: <i>15:30 30.03.2025</i>\n" +
		"<blockquote>This is the body of the pull request. It contains details about the changes.</blockquote>\n"
	assert.Equal(t, exp, msg, "invalid message")
}

func TestDataToMessage_Issue(t *testing.T) {
	data := github.Data{
		PR:        github.PullRequest{URL: ""},
		URL:       "https://github.com/example/repo/issues/42",
		Title:     "Bug in feature",
		User:      github.User{Login: "bugHunter"},
		CreatedAt: time.Date(2025, 4, 1, 10, 0, 0, 0, time.Local),
		Body:      strings.Repeat("a", 250),
	}

	msg := github.DataToMessage(&data)

	exp := "<b>Новое Issue на Github!</b>\n" +
		"<b>Название</b>: <a href=\"https://github.com/example/repo/issues/42\">Bug in feature</a>\n" +
		"<b>Автор</b>: <i>bugHunter</i>\n" +
		"<b>Время создания</b>: <i>10:00 01.04.2025</i>\n" +
		fmt.Sprintf("<blockquote>%s</blockquote>\n", strings.Repeat("a", 200)+"...")
	assert.Equal(t, exp, msg, "invalid message")
}
