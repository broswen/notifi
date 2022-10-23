package main

import (
	"context"
	"fmt"
	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/destination"
	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/queue/consumer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
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

	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "8081"
	}

	metricsPath := os.Getenv("METRICS_PATH")
	if metricsPath == "" {
		metricsPath = "/metrics"
	}

	fromName := os.Getenv("FROM_NAME")
	if fromName == "" {
		log.Fatal().Msgf("from email name is empty")
	}

	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		log.Fatal().Msgf("from email is empty")
	}

	sendGridApiKey := os.Getenv("SENDGRID_KEY")
	if sendGridApiKey == "" {
		log.Fatal().Msgf("sendgrid api key is empty")
	}

	twilioAccountSid := os.Getenv("TWILIO_SID")
	if twilioAccountSid == "" {
		log.Fatal().Msgf("twilio account sid is empty")
	}

	twilioAuthToken := os.Getenv("TWILIO_TOKEN")
	if twilioAuthToken == "" {
		log.Fatal().Msgf("twilio auth token is empty")
	}

	fromNumber := os.Getenv("FROM_NUMBER")
	if fromNumber == "" {
		log.Fatal().Msgf("from number is empty")
	}

	skipDelivery := os.Getenv("SKIP_DELIVERY")
	if skipDelivery != "" {
		log.Debug().Str("SKIP_DELIVERY", skipDelivery).Msg("skip delivery mode")
	}

	email, err := destination.NewEmailDestination(sendGridApiKey, fromName, fromEmail)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating email destination")
	}

	sms, err := destination.NewSMSDestination(twilioAccountSid, twilioAuthToken, fromNumber)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating sms destination")
	}

	l, err := destination.NewLogDestination("delivery")
	if err != nil {
		log.Fatal().Err(err).Msg("error creating log destination")
	}

	c, err := consumer.NewKafkaConsumer("delivery", "delivery", deliveryTopic, brokers)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating kafka consumer")
	}

	pool, err := db.InitDB(context.Background(), dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating postgres pool")
	}
	notificationRepo, err := repository.NewNotificationSqlRepository(pool)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating notification repository")
	}

	c.HandleFunc(deliveryTopic, func(n entity.Notification) error {
		var err error
		if skipDelivery != "" {
			err = l.Deliver(n)
		} else {
			if n.Destination.Email != "" {
				err = email.Deliver(n)
			} else if n.Destination.SMS != "" {
				err = sms.Deliver(n)
			} else {
				err = fmt.Errorf("notification missing destination: %s", n.ID)
				log.Error().Err(err).Str("notification_id", n.ID).Msg("notification missing destination")
			}
			log.Error().Err(err).Str("notification_id", n.ID).Msg("notification delivery error")
			return err
		}

		now := time.Now()
		n.DeliveredAt = &now
		_, err = notificationRepo.Update(context.Background(), n)
		if err != nil {
			//TODO add something to prevent frequent delivery retries if the delivery succeeds but database fails to save
			log.Error().Err(err).Msg("error updating notification")
		}
		if !n.Scheduled() {
			//DeliveryDelay is the time from creation to delivery for instant notifications
			DeliveryDelay.Observe(float64(n.DeliveredAt.Sub(n.CreatedAt)))
		} else {
			//DeliveryDelay is the difference from scheduled time to delivery for scheduled notifications
			//a negative value means it was delivered early
			ScheduledOffset.Observe(float64(n.DeliveredAt.Sub(*n.Schedule)))
		}
		return err
	})

	err = c.Consume()
	if err != nil {
		log.Fatal().Err(err).Msg("error starting kafka consumer")
	}
	defer c.Close()

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

	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("")
	}
}
