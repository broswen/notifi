package poller

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SuccessfulPoll = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poll_success",
	})

	PollErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poll_error",
	})

	PollNotifications = promauto.NewCounter(prometheus.CounterOpts{
		Name: "poll_notifications",
	})
)
