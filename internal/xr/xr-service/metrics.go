package xrservice

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	externalRequestDuration *prometheus.HistogramVec
}

const (
	namespace = "xr_service"
	subsystem = "service"
)

func newMetric() *metrics {
	metricList := metrics{
		externalRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "external_request_duration_seconds",
				Help:      "Duration of external HTTP requests.",
			},
			[]string{"endpoint"}),
	}

	return &metricList
}
