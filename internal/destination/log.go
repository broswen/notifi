package destination

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/rs/zerolog/log"
)

type LogDestination struct {
	name string
}

func (d *LogDestination) Deliver(n entity.Notification) error {
	log.Debug().Str("name", d.name).Str("notification_id", n.ID).Msgf("delivered %s to %s", n.ID, d.name)
	return nil
}
