package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/broswen/notifi/internal/entity"
	"github.com/rs/zerolog/log"
	log2 "log"
	"os"
	"strings"
)

type KafkaConsumer struct {
	id       string
	group    string
	topics   string
	brokers  string
	client   sarama.ConsumerGroup
	cancel   context.CancelFunc
	handlers map[string]NotificationHandler
}

func NewKafkaConsumer(id, group, topics, brokers string) (Consumer, error) {

	config := sarama.NewConfig()
	config.ClientID = id
	version, err := sarama.ParseKafkaVersion("3.1.0")
	if err != nil {
		return nil, err
	}
	config.Version = version

	sarama.Logger = log2.New(os.Stdout, "[sarama] ", log2.LstdFlags)

	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		return nil, err
	}

	c := &KafkaConsumer{
		id:       id,
		group:    group,
		topics:   topics,
		brokers:  brokers,
		client:   client,
		handlers: make(map[string]NotificationHandler),
	}

	return c, nil
}

func (c *KafkaConsumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}

func (c *KafkaConsumer) Consume() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go func() {
		for {
			if err := c.client.Consume(ctx, strings.Split(c.topics, ","), c); err != nil {
				log.Panic().Err(err)
			}
			if err := ctx.Err(); err != nil {
				log.Error().Err(err).Msg("")
				return
			}
		}
	}()
	return nil
}

func (c *KafkaConsumer) HandleFunc(topic string, h NotificationHandler) {
	c.handlers[topic] = h
}

func (c *KafkaConsumer) Handle(message *sarama.ConsumerMessage) error {
	if h, ok := c.handlers[message.Topic]; ok {
		notification := entity.Notification{}
		err := json.Unmarshal(message.Value, &notification)
		if err != nil {
			return err
		}
		return h(notification)
	}
	return errors.New("no matching handler: " + message.Topic)
}

func (c *KafkaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *KafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			err := c.Handle(message)
			if err != nil {
				log.Error().Err(err)
				continue
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
