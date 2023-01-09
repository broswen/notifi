package main

import (
	"context"
	"fmt"
	"github.com/broswen/notifi/internal/destination"
	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/queue/consumer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"
	"time"
)

func HandleDelivery(notificationRepo repository.NotificationRepository, sms destination.Destination, email destination.Destination, l destination.Destination, rdb *redis.Client, redisTTL time.Duration, skipDelivery string) consumer.NotificationHandler {
	return func(n entity.Notification) error {
		if rdb != nil {
			_, err := rdb.Get(context.Background(), n.ID).Result()
			if err != nil {
				// either redis lookup failed, or key doesn't exist
				if err != redis.Nil {
					// if lookup failed, log error
					log.Error().Err(err).Msg("error getting notification id from redis")
				}
			} else {
				//no error means the key exists and has been delivered already
				log.Warn().Str("notification_id", n.ID).Msg("duplicate notification id skipped")
				DuplicateDelivery.Inc()
				return nil
			}
		}

		//check DB is up before trying to deliver messages
		err := notificationRepo.Ping(context.Background())
		if err != nil {
			return err
		}
		if skipDelivery != "" {
			//artificial delay to mimic network request
			time.Sleep(time.Millisecond * 300)
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
			if err != nil {
				log.Error().Err(err).Str("notification_id", n.ID).Msg("notification delivery error")
				return err
			}
		}

		now := time.Now()
		n.DeliveredAt = &now
		_, err = notificationRepo.Update(context.Background(), n)
		if err != nil {
			//TODO add something to prevent frequent delivery retries if the delivery succeeds but database fails to save
			log.Error().Err(err).Msg("error updating notification")
			return err
		}

		if rdb != nil {
			// mark notification id as seen
			_, err = rdb.Set(context.Background(), n.ID, true, redisTTL).Result()
			if err != nil {
				log.Error().Err(err).Str("notification_id", n.ID).Msg("error setting notification id in redis")
			}
		}

		if !n.Scheduled() {
			//DeliveryDelay is the time from creation to delivery for instant notifications
			DeliveryDelay.Observe(float64(n.DeliveredAt.Sub(n.CreatedAt)))
		} else {
			//DeliveryDelay is the difference from scheduled time to delivery for scheduled notifications
			//a negative value means it was delivered early
			ScheduledOffset.Observe(float64(n.DeliveredAt.Sub(*n.Schedule)))
		}
		return nil
	}

}
