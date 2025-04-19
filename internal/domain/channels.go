package domain

import (
	"github.com/es-debug/backend-academy-2024-go-template/pkg/kafka"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Channels struct {
	telegramReq  chan TelegramRequest
	telegramResp chan tgbotapi.Chattable
	kafkaInput   chan *kafka.Input
	kafkaOutput  chan *kafka.Message
}

func NewChannels() *Channels {
	return &Channels{
		telegramReq:  make(chan TelegramRequest),
		telegramResp: make(chan tgbotapi.Chattable),
		kafkaInput:   make(chan *kafka.Input),
		kafkaOutput:  make(chan *kafka.Message),
	}
}

func (c *Channels) TelegramReq() chan TelegramRequest {
	return c.telegramReq
}

func (c *Channels) TelegramResp() chan tgbotapi.Chattable {
	return c.telegramResp
}

func (c *Channels) KafkaInput() chan *kafka.Input {
	return c.kafkaInput
}

func (c *Channels) KafkaOutput() chan *kafka.Message {
	return c.kafkaOutput
}
