package producer

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/broswen/notifi/internal/entity"
	"github.com/rs/zerolog/log"
	"time"
)

type KafkaProducer struct {
	id       string
	broker   string
	topic    string
	producer sarama.SyncProducer
}

func NewKafkaProducer(id, topic, broker string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.ClientID = id
	version, err := sarama.ParseKafkaVersion("3.1.0")
	if err != nil {
		return nil, err
	}
	config.Version = version
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{
		id:       id,
		broker:   broker,
		topic:    topic,
		producer: producer,
	}, nil
}

func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

func (p *KafkaProducer) Submit(notification entity.Notification) error {
	j, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       sarama.StringEncoder(notification.ID),
		Value:     sarama.StringEncoder(j),
		Timestamp: time.Now(),
	})
	if err != nil {
		log.Err(err).Str("producer_id", p.id).Str("topic", p.topic).Str("broker", p.broker).Str("notification_id", notification.ID).Msg("couldn't submit notification")
		return err
	}
	return nil
}
