package main

import (
	"context"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func main() {

	// parse queue topic
	// parse brokers
	// parse dsn
	// parse poll interval
	brokers := os.Getenv("BROKERS")
	if brokers == "" {
		log.Fatal().Msgf("kafka brokers are empty")
	}
	deliveryTopic := os.Getenv("DELIVERY_TOPIC")
	if deliveryTopic == "" {
		log.Fatal().Msgf("delivery topic is empty")
	}
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal().Msgf("postgres DSN is empty")
	}
	pollInterval := os.Getenv("POLL_INTERVAL")
	if pollInterval == "" {
		log.Fatal().Msgf("poll interval is empty")
	}
	interval, err := time.ParseDuration(pollInterval)
	if err != nil {
		log.Fatal().Err(err).Msg("error parsing poll interval")
	}

	p, err := producer.NewKafkaProducer("poller", deliveryTopic, brokers)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating kafka producer")
	}

	ticker := time.NewTicker(interval)
	ctx, cancel := context.WithCancel(context.Background())

	for {
		select {
		case <-ticker.C:
			log.Debug().Msg("polling for scheduled messages")
		case <-ctx.Done():

		}
	}
	// scan db for notifications where the scheduled time is less than X minutes in the future
	// submit to queue
	// remove from db (or mark sent for a cleanup job)
	// store successful notifications in postgres
	//		store failed notification status
}
