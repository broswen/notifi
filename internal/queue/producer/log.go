package producer

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/rs/zerolog/log"
)

type LogProducer struct {
	id    string
	topic string
}

func NewLogProducer(id, topic string) (Producer, error) {
	return &LogProducer{
		id:    id,
		topic: topic,
	}, nil
}

func (p *LogProducer) Close() error {
	return nil
}

func (p *LogProducer) Submit(notification entity.Notification) error {
	log.Debug().Str("producer_id", p.id).Str("topic", p.topic).Str("notification_id", notification.ID).Msgf("submitted notification %s to %s", notification.ID, p.topic)
	return nil
}
