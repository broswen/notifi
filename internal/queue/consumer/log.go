package consumer

import (
	"fmt"
	"github.com/broswen/notifi/internal/entity"
	"github.com/rs/zerolog/log"
)

type LogConsumer struct {
	name     string
	handlers map[string]NotificationHandler
}

func NewLogConsumer(name string) Consumer {
	return &LogConsumer{
		name:     name,
		handlers: make(map[string]NotificationHandler),
	}
}

func (c *LogConsumer) Close() error {
	return nil
}
func (c *LogConsumer) Consume() error {
	return nil
}

func (c *LogConsumer) HandleFunc(topic string, h NotificationHandler) {
	c.handlers[topic] = h
}

// Submit is used for testing
func (c *LogConsumer) Submit(topic string, n entity.Notification) error {
	if h, ok := c.handlers[topic]; ok {
		err := h(n)
		if err != nil {
			log.Error().Err(err).Str("name", c.name).Str("topic", topic).Msg("")
		}
		return err
	}
	return fmt.Errorf("no handler for topic: %s", topic)
}
