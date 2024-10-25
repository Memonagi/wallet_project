package application

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	txFailed        *prometheus.CounterVec
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func newMetrics() *metrics {
	metricList := metrics{
		txFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tx_failed_total",
				Help: "Number of failed transactions.",
			},
			[]string{"endpoint"}),
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_request_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"endpoint"}),

		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_request_duration_seconds",
				Help: "Duration of HTTP requests.",
			},
			[]string{"endpoint"}),
	}

	prometheus.MustRegister(
		metricList.txFailed,
		metricList.requestsTotal,
		metricList.requestDuration,
	)

	return &metricList
}

func (m *metrics) TrackHTTPRequest(start time.Time, r *http.Request) {
	id := chi.URLParam(r, "id")
	url := r.URL.Host + r.URL.Path

	if id != "" {
		url = strings.Replace(url, id, "{id}", 1)
	}

	timePassed := time.Since(start).Seconds()

	m.requestsTotal.WithLabelValues(r.Method + url).Inc()
	m.requestDuration.WithLabelValues(r.Method + url).Observe(timePassed)
}
