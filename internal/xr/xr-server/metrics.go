package xrserver

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	externalRequestDuration *prometheus.HistogramVec
}

const (
	namespace = "xr_service"
	subsystem = "server"
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

func (m *metrics) trackExternalRequest(start time.Time, endpoint string) {
	timePassed := time.Since(start).Seconds()

	m.externalRequestDuration.WithLabelValues(endpoint).Observe(timePassed)
}
