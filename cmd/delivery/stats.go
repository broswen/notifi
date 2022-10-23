package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"time"
)

var (
	DeliveryDelay = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "delivery_delay",
		Buckets: []float64{float64(time.Second * 10), float64(time.Second * 30), float64(time.Minute), float64(time.Minute * 3), float64(time.Minute * 6), float64(time.Minute * 10)},
	})
	ScheduledOffset = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "scheduled_delivery_offset",
		Buckets: []float64{float64(time.Minute * -10), float64(time.Minute * -6), float64(time.Minute * -3), float64(time.Minute * -1), float64(time.Second * -30), float64(time.Second * -10)},
	})
)
