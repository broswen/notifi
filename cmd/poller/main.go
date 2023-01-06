package main

import (
	"context"
	"fmt"
	"github.com/broswen/notifi/internal/db"
	poller2 "github.com/broswen/notifi/internal/poller"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "8081"
	}

	metricsPath := os.Getenv("METRICS_PATH")
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	pollInterval := os.Getenv("POLL_INTERVAL")
	if pollInterval == "" {
		log.Fatal().Msgf("poll interval is empty")
	}
	interval, err := time.ParseDuration(pollInterval)
	if err != nil {
		log.Fatal().Err(err).Msg("error parsing poll interval")
	}

	partitionStart := os.Getenv("PARTITION_START")
	var partitionStartValue	int64
	if partitionStart != "" {
		partitionStartValue, err = strconv.ParseInt(partitionStart, 10, 64)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid PARTITION_START")
		}
	} else {
		log.Fatal().Msg("PARTITION_START must be defined")
	}


	partitionEnd := os.Getenv("PARTITION_END")
	var partitionEndValue int64
	if partitionEnd != "" {
		partitionEndValue, err = strconv.ParseInt(partitionEnd, 10, 64)
		if err != nil {
			log.Fatal().Err(err).Msg("invalid PARTITION_END")
		}
	} else {
		log.Fatal().Msg("PARTITION_END must be defined")
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

	ctx, cancel := context.WithCancel(context.Background())
	pollPeriod := time.Minute * 5
	pollLimit := int64(100)
	poller := poller2.NewScheduledNotificationPoller(scheduledRepo, p, interval, pollPeriod, pollLimit, partitionStartValue, partitionEndValue)

	eg.Go(func() error {
		return poller.Poll(ctx)
	})

	m := chi.NewRouter()
	m.Handle(metricsPath, promhttp.Handler())
	eg.Go(func() error {
		if err := http.ListenAndServe(fmt.Sprintf(":%s", metricsPort), m); err != nil {
			if err != http.ErrServerClosed {
				return err
			}
		}
		return nil
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
