package poller

import (
	"context"
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
)

// Poller is an interface that polls and submits the items to a producer
type Poller interface {
	Poll(context.Context) error
	Submit(context.Context, entity.Notification) error
}

type ScheduledNotificationPoller struct {
	db           *sql.DB
	Producer     producer.Producer
	pollInterval time.Duration
	pollPeriod   time.Duration
	pollLimit    int64
}

func NewScheduledNotificationPoller(db *sql.DB, producer producer.Producer, pollInterval, pollPeriod time.Duration, pollLimit int64) *ScheduledNotificationPoller {
	return &ScheduledNotificationPoller{
		db:           db,
		Producer:     producer,
		pollInterval: pollInterval,
		pollPeriod:   pollPeriod,
		pollLimit:    pollLimit,
	}
}

// Poll polls Notification repository every p.pollInterval for p.pollLimit notifications that are due within p.PollPeriod
// it will attempt to delivery every notification to p.Destination
func (p *ScheduledNotificationPoller) Poll(ctx context.Context) error {
	ticker := time.NewTicker(p.pollInterval)
	for {
		select {
		case <-ticker.C:
			p.poll(ctx)
		case <-ctx.Done():
			log.Debug().Msg("context cancelled")
			return nil
		}
	}
}

func (p *ScheduledNotificationPoller) poll(ctx context.Context) error {
	extras := true
	offset := int64(0)
	for extras {
		log.Debug().
			Str("interval", p.pollInterval.String()).
			Int64("limit", p.pollLimit).
			Msg("polling for scheduled messages")
		tx, err := p.db.BeginTx(ctx, nil)
		if err != nil {
			log.Error().Err(err).Msg("error starting tx")
			return err
		}
		notifications, err := repository.NewScheduledNotificationSqlRepository(tx).ListScheduled(ctx, p.pollPeriod, p.pollLimit)
		if err != nil {
			log.Error().Err(err).Msg("error listing scheduled notifications")
			PollErrors.Inc()
			break
		}
		PollNotifications.Add(float64(len(notifications)))
		//continue loop if we received a full amount
		//offset next query
		extras = int64(len(notifications)) == p.pollLimit
		offset += p.pollLimit
		for _, n := range notifications {
			err := p.Submit(ctx, tx, n)
			if err != nil {
				log.Error().Err(err).Str("notification_id", n.ID).Msg("error submitting notification")
				continue
			}
		}
		if err := tx.Commit(); err != nil {
			log.Error().Err(err).Msg("error ending tx")
		}
		SuccessfulPoll.Inc()
	}
	return nil
}

func (p *ScheduledNotificationPoller) Submit(ctx context.Context, db db.Conn, n entity.Notification) error {
	log.Debug().Str("notification_id", n.ID).Time("schedule", *n.Schedule).Msg("submitting scheduled notification")
	err := p.Producer.Submit(n)
	if err != nil {
		return err
	}
	now := time.Now()
	n.SubmittedAt = &now
	_, err = repository.NewScheduledNotificationSqlRepository(db).MarkSubmitted(ctx, n.ID)
	if err != nil {
		log.Error().Err(err).Str("notification_id", n.ID).Msg("error marking notification as submitted")
		return err
	}
	return nil
}
