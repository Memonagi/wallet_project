package xrservice

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	externalRequestDuration *prometheus.HistogramVec
}

func newMetrics() *metrics {
	metricList := metrics{
		externalRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "external_request_duration_seconds",
				Help: "Duration of external HTTP requests.",
			},
			[]string{"endpoint"}),
	}

	prometheus.MustRegister(metricList.externalRequestDuration)

	return &metricList
}

func (m *metrics) TrackExternalRequest(start time.Time, endpoint string) {
	timePassed := time.Since(start).Seconds()

	m.externalRequestDuration.WithLabelValues(endpoint).Observe(timePassed)
}
