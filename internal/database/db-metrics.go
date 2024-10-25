package database

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	txDuration *prometheus.HistogramVec
}

func newMetric() *metrics {
	metric := metrics{
		txDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "tx_duration_seconds",
				Help: "Duration of transaction.",
			},
			[]string{"endpoint"}),
	}

	prometheus.MustRegister(metric.txDuration)

	return &metric
}
