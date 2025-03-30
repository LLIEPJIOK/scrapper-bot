package github

import (
	"fmt"
	"strings"
)

func DataToMessage(data *Data) string {
	builder := strings.Builder{}
	if data.PR.URL != "" {
		builder.WriteString("<b>Новый Pull Request на Github!</b>\n")
	} else {
		builder.WriteString("<b>Новое Issue на Github!</b>\n")
	}

	builder.WriteString(
		fmt.Sprintf("<b>Название</b>: <a href=%q>%s</a>\n", data.URL, data.Title),
	)
	builder.WriteString(fmt.Sprintf("<b>Автор</b>: <i>%s</i>\n", data.User.Login))
	builder.WriteString(
		fmt.Sprintf(
			"<b>Время создания</b>: <i>%s</i>\n",
			data.CreatedAt.Local().Format("15:04 02.01.2006"),
		),
	)
	builder.WriteString(fmt.Sprintf("<blockquote>%s</blockquote>\n", preview(data.Body)))

	return builder.String()
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
