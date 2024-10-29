package database

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	txDuration *prometheus.HistogramVec
}

const (
	namespace = "wallet_service"
	subsystem = "database"
)

func newMetric() *metrics {
	metric := metrics{
		txDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_duration_seconds",
				Help:      "Duration of transaction.",
			},
			[]string{"endpoint"}),
	}

	return &metric
}
