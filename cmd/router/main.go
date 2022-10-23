package main

import (
	"fmt"
	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/queue/consumer"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	brokers := os.Getenv("BROKERS")
	if brokers == "" {
		log.Fatal().Msgf("kafka brokers are empty")
	}
	requestTopic := os.Getenv("REQUEST_TOPIC")
	if requestTopic == "" {
		log.Fatal().Msgf("request topic is empty")
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

	c, err := consumer.NewKafkaConsumer("router", "router", requestTopic, brokers)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating kafka consumer")
	}
	p1, err := producer.NewKafkaProducer("router", deliveryTopic, brokers)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating kafka producer")
	}
	defer p1.Close()

	p2, err := producer.NewLogProducer("router", "scheduled")
	if err != nil {
		log.Fatal().Err(err).Msg("error creating log producer")
	}
	defer p2.Close()

	eg := errgroup.Group{}
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

	c.HandleFunc(requestTopic, func(n entity.Notification) error {
		if n.Scheduled() {
			//TODO add postgres producer to store notifications
			return p2.Submit(n)
			return nil
		} else {
			//submit to delivery queue if instant notification
			return p1.Submit(n)
		}
	})

	err = c.Consume()
	if err != nil {
		log.Fatal().Err(err).Msg("error starting kafka consumer")
	}
	defer c.Close()

	eg.Go(func() error {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		s := <-sigs
		log.Debug().Str("signal", s.String()).Msg("received signal")
		if err := c.Close(); err != nil {
			log.Error().Err(err).Msg("error closing kafka consumer")
			return err
		}
		return nil
	})

	if err = eg.Wait(); err != nil {
		log.Error().Err(err).Msg("")
	}
}
