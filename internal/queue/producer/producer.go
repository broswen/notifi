package producer

import "github.com/broswen/notifi/internal/entity"

type Producer interface {
	Submit(notification entity.Notification) error
	Close() error
}

func NewProducer(id, topic, broker string) (Producer, error) {
	if broker == "" || broker == "logger" {
		return NewLogProducer(id, topic)
	} else {
		return NewKafkaProducer(id, topic, broker)
	}
}
