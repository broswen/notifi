package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/broswen/notifi/internal/api"
	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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

	partitions := os.Getenv("PARTITIONS")
	partitionsValue	:= 100
	if partitions != "" {
		val, err := strconv.ParseInt(partitions, 10, 32)
		if err != nil {
			log.Fatal().Err(err)
		}
		partitionsValue = int(val)
	}

	//Cloudflare Access Application policy AUD
	policyAUD := os.Getenv("POLICY_AUD")
	//Cloudflare Access team domain <team>.cloudflareaccess.com
	teamDomain := os.Getenv("TEAM_DOMAIN")

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

	pool, err := db.InitDB(context.Background(), dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating postgres pool")
	}
	notificationRepo, err := repository.NewNotificationSqlRepository(pool)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating notification repository")
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

	app := api.API{
		Partitions: partitionsValue,
		Producer:     p1,
		Notification: notificationRepo,
	}

	accessClient := api.NewAccessClient(teamDomain, policyAUD)
	publicServer := http.Server{
		Addr:    fmt.Sprintf(":%s", apiPort),
		Handler: app.Router(accessClient),
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
		sig := <-sigint
		log.Debug().Str("signal", sig.String()).Msg("received signal")
		if err := publicServer.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("error shutting down public server")
		}
		return err
	})

	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("")
	}
}
