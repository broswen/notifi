package main

import (
	"context"
	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

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

	pool, err := db.InitDB(context.Background(), dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating postgres pool")
	}
	scheduledRepo, err := repository.NewScheduledNotificationSqlRepository(pool)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating scheduled notification repository")
	}
	eg := errgroup.Group{}

	ticker := time.NewTicker(interval)
	ctx, cancel := context.WithCancel(context.Background())

	//notifications can be delivered up to 5 minutes early
	pollPeriod := time.Minute * 5
	pollLimit := int64(100)
	eg.Go(func() error {
		for {
			select {
			case <-ticker.C:
				log.Debug().Str("interval", interval.String()).Int64("limit", pollLimit).Msg("polling for scheduled messages")
				notifications, err := scheduledRepo.ListScheduled(ctx, pollPeriod, pollLimit)
				if err != nil {
					log.Error().Err(err).Msg("error listing scheduled notifications")
					continue
				}
				for _, n := range notifications {
					//TODO mark in-progress to avoid resubmitting during another poll
					log.Debug().Str("notification_id", n.ID).Time("schedule", *n.Schedule).Msg("submitting scheduled notification")
					err = p.Submit(n)
					if err != nil {
						log.Error().Err(err).Str("notification_id", n.ID).Msg("error submitting scheduled notification")
						continue
					}
				}
			case <-ctx.Done():
				log.Debug().Msg("context cancelled")
				return nil
			}
		}
	})

	eg.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		s := <-sigs
		log.Debug().Str("signal", s.String()).Msg("received signal")
		cancel()
		if err := p.Close(); err != nil {
			log.Error().Err(err).Msg("error closing kafka producer")
			return err
		}
		return nil
	})

	if err = eg.Wait(); err != nil {
		log.Error().Err(err).Msg("")
	}
}
