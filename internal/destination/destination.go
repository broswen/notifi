package destination

import "github.com/broswen/notifi/internal/entity"

//a destination receives a message attempts delivery and retries

type Destination interface {
	Deliver(notification entity.Notification) error
}