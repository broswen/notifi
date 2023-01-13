package destination

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/rs/zerolog/log"
)

type LogDestination struct {
	name string
}

func NewLogDestination(name string) (Destination, error) {
	return &LogDestination{
		name: name,
	}, nil
}

func (d *LogDestination) Deliver(n entity.Notification) error {
	log.Debug().Str("name", d.name).Str("notification_id", n.ID).Msgf("delivered %s to %s", n.ID, d.name)
	NotificationDelivered.WithLabelValues("log").Inc()
	return nil
}
