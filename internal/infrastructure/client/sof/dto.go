package sof

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

func AnswersToMessages(answers []Answer, questionTitle, questionLink string) []string {
	msgs := make([]string, 0, len(answers))

	for _, answer := range answers {
		builder := strings.Builder{}

		builder.WriteString("<b>Новый ответ на StackOverflow!</b>\n")
		builder.WriteString(
			fmt.Sprintf(
				"<b>Вопрос</b>: <a href=%q>%s</a>\n",
				questionLink,
				clearHTML(questionTitle),
			),
		)
		builder.WriteString(fmt.Sprintf("<b>Автор</b>: <i>%s</i>\n", answer.Owner.DisplayName))
		builder.WriteString(
			fmt.Sprintf(
				"<b>Время создания</b>: <i>%s</i>\n",
				time.Unix(answer.CreatedAt, 0).Format("15:04 02.01.2006"),
			),
		)
		builder.WriteString(
			fmt.Sprintf("<blockquote>%s</blockquote>\n", preview(clearHTML(answer.Body))),
		)

		msgs = append(msgs, builder.String())
	}

	return msgs
}

func CommentsToMessages(comment []Comment, questionTitle, questionLink string) []string {
	msgs := make([]string, 0, len(comment))

	for _, com := range comment {
		builder := strings.Builder{}

		builder.WriteString("<b>Новый комментарий на StackOverflow!</b>\n")
		builder.WriteString(
			fmt.Sprintf(
				"<b>Вопрос</b>: <a href=%q>%s</a>\n",
				questionLink,
				clearHTML(questionTitle),
			),
		)
		builder.WriteString(fmt.Sprintf("<b>Автор</b>: <i>%s</i>\n", com.Owner.DisplayName))
		builder.WriteString(
			fmt.Sprintf(
				"<b>Время создания</b>: <i>%s</i>\n",
				time.Unix(com.CreatedAt, 0).Format("15:04 02.01.2006"),
			),
		)
		builder.WriteString(
			fmt.Sprintf("<blockquote>%s</blockquote>\n", preview(clearHTML(com.Body))),
		)

		msgs = append(msgs, builder.String())
	}

	return msgs
}

func preview(text string) string {
	rns := []rune(text)
	if len(rns) > 200 {
		rns = rns[:200]
		rns = append(rns, []rune("...")...)

		return string(rns)
	}

	return text
}

func clearHTML(text string) string {
	text = strings.TrimSpace(text)

	re := regexp.MustCompile(`<[^>]+>`)

	return re.ReplaceAllString(text, "")
}
