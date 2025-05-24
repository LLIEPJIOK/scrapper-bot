package kafka

type MessageChannels struct {
	ack  chan *Message
	nack chan *Message
	dlq  chan *Message
}

func NewMessageChannels() *MessageChannels {
	return &MessageChannels{
		ack:  make(chan *Message),
		nack: make(chan *Message),
		dlq:  make(chan *Message),
	}
}

func (c *MessageChannels) Close() {
	close(c.ack)
	close(c.nack)
	close(c.dlq)
}

func (c *MessageChannels) Ack() chan *Message {
	return c.ack
}

func (c *MessageChannels) Nack() chan *Message {
	return c.nack
}

func (c *MessageChannels) DLQ() chan *Message {
	return c.dlq
}
