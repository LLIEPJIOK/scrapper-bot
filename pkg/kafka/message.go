package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

type Message struct {
	Base       *sarama.ConsumerMessage
	retryCount int32

	channels *MessageChannels
}

func NewMessage(msg *sarama.ConsumerMessage, retryCount int32, channels *MessageChannels) *Message {
	return &Message{
		Base:       msg,
		retryCount: retryCount,
		channels:   channels,
	}
}

func (m *Message) Bind(v any) error {
	return json.Unmarshal(m.Base.Value, v)
}

func (m *Message) Ack() {
	if m.retryCount == 0 {
		m.channels.Ack() <- m
	}
}

func (m *Message) Nack() {
	m.retryCount++
	m.channels.Nack() <- m
}

func (m *Message) NackToDLQ() {
	m.channels.DLQ() <- m
}

func (m *Message) RetryCount() int32 {
	return m.retryCount
}

type Input struct {
	Topic string
	Key   string
	Value string
}
