package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NotificationDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_deleted",
	})
	NotificationCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "notification_created",
	})
)
