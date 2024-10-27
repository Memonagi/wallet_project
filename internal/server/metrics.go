package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

const (
	namespace = "wallet_service"
	subsystem = "server"
)

func newMetrics() *metrics {
	metricList := metrics{
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_request_total",
				Help:      "Total number of HTTP requests.",
			},
			[]string{"endpoint"}),

		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests.",
			},
			[]string{"endpoint"}),
	}

	return &metricList
}

func (m *metrics) trackHTTPRequest(start time.Time, r *http.Request) {
	id := chi.URLParam(r, "id")
	url := r.URL.Host + r.URL.Path

	if id != "" {
		url = strings.Replace(url, id, "{id}", 1)
	}

	timePassed := time.Since(start).Seconds()

	m.requestsTotal.WithLabelValues(r.Method + url).Inc()
	m.requestDuration.WithLabelValues(r.Method + url).Observe(timePassed)
}
