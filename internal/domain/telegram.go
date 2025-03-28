package domain

const (
	Message  TgReqType = "message"
	Command  TgReqType = "command"
	Callback TgReqType = "callback"
)

type TgReqType string

type TelegramRequest struct {
	ChatID    int64
	MessageID int
	Message   string
	Type      TgReqType
}
