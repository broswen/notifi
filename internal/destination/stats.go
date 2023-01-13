package destination

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NotificationDelivered = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_delivered",
	}, []string{"destination_type"})
)
