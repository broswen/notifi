package main

import (
	"context"
	"fmt"
	"github.com/broswen/notifi/internal/api"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		apiPort = "8080"
	}

	brokers := os.Getenv("BROKERS")
	if brokers == "" {
		log.Fatal().Msgf("kafka brokers are empty")
	}
	requestTopic := os.Getenv("REQUEST_TOPIC")
	if requestTopic == "" {
		log.Fatal().Msgf("request topic is empty")
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

	p1, err := producer.NewKafkaProducer("router", requestTopic, brokers)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating kafka producer")
	}

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

	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Post("/", api.HandleCreateNotification(p1))
	publicServer := http.Server{
		Addr:    fmt.Sprintf(":%s", apiPort),
		Handler: r,
	}
	eg.Go(func() error {
		log.Debug().Msgf("public api listening on :%s", apiPort)
		if err := publicServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		if err := publicServer.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("error shutting down public server")
		}
		return err
	})

	err = eg.Wait()
	if err != nil {
		log.Error().Err(err).Msg("")
	}
}
