package application

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	txFailed    *prometheus.CounterVec
	txCompleted *prometheus.CounterVec
}

const (
	namespace = "wallet_service"
	subsystem = "application"
)

func newMetrics() *metrics {
	metricList := metrics{
		txFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_failed_total",
				Help:      "Number of failed transactions.",
			},
			[]string{"endpoint"}),
		txCompleted: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "tx_completed_total",
				Help:      "Number of completed transactions.",
			},
			[]string{"endpoint"}),
	}

	return &metricList
}
