package domain

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Channels struct {
	telegramReq  chan TelegramRequest
	telegramResp chan tgbotapi.Chattable
}

func NewChannels() *Channels {
	return &Channels{
		telegramReq:  make(chan TelegramRequest),
		telegramResp: make(chan tgbotapi.Chattable),
	}
}

func (c *Channels) TelegramReq() chan TelegramRequest {
	return c.telegramReq
}

func (c *Channels) TelegramResp() chan tgbotapi.Chattable {
	return c.telegramResp
}
