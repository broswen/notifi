package consumer

import (
	"github.com/broswen/notifi/internal/entity"
)

//consumer should take an id, broker, list of topics
//register handlers for each topic

type Consumer interface {
	HandleFunc(topic string, h NotificationHandler)
	Consume() error
	Close() error
}

type NotificationHandler func(n entity.Notification) error
