package poller

import (
	"context"
	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/queue/producer"
	"github.com/broswen/notifi/internal/repository"
	"github.com/rs/zerolog/log"
	"time"
)

// Poller is an interface that polls and submits the items to a producer
type Poller interface {
	Poll(context.Context) error
	Submit(context.Context, entity.Notification) error
}

type ScheduledNotificationPoller struct {
	Notification repository.ScheduledNotificationRepository
	Producer     producer.Producer
	pollInterval time.Duration
	pollPeriod   time.Duration
	pollLimit    int64
	partitionStart int64
	partitionEnd int64
}

func NewScheduledNotificationPoller(notificationRepository repository.ScheduledNotificationRepository, producer producer.Producer, pollInterval, pollPeriod time.Duration, pollLimit, partitionStart, partitionEnd int64) *ScheduledNotificationPoller {
	return &ScheduledNotificationPoller{
		Notification: notificationRepository,
		Producer:     producer,
		pollInterval: pollInterval,
		pollPeriod:   pollPeriod,
		pollLimit:    pollLimit,
		partitionStart: partitionStart,
		partitionEnd: partitionEnd,
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

func (p *ScheduledNotificationPoller) poll(ctx context.Context) {
	extras := true
	offset := int64(0)
	for extras {
		log.Debug().
			Str("interval", p.pollInterval.String()).
			Int64("limit", p.pollLimit).
			Int64("partition_start", p.partitionStart).
			Int64("partition_end", p.partitionEnd).
			Msg("polling for scheduled messages")
		//TODO add notification partition key? for polling so we can scale the poller
		notifications, err := p.Notification.ListScheduled(ctx, p.pollPeriod, p.partitionStart, p.partitionEnd, offset, p.pollLimit)
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
			err := p.Submit(ctx, n)
			if err != nil {
				log.Error().Err(err).Str("notification_id", n.ID).Msg("error submitting notification")
				continue
			}
		}
		SuccessfulPoll.Inc()
	}
}

func (p *ScheduledNotificationPoller) Submit(ctx context.Context, n entity.Notification) error {
	log.Debug().Str("notification_id", n.ID).Time("schedule", *n.Schedule).Msg("submitting scheduled notification")
	err := p.Producer.Submit(n)
	if err != nil {
		return err
	}
	now := time.Now()
	n.SubmittedAt = &now
	_, err = p.Notification.MarkSubmitted(ctx, n.ID)
	if err != nil {
		log.Error().Err(err).Str("notification_id", n.ID).Msg("error marking notification as submitted")
		return err
	}
	return nil
}
