package xrserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type service interface {
	GetRate(request models.XRRequest) (float64, error)
}

type metrics interface {
	TrackExternalRequest(start time.Time, endpoint string)
}

type Server struct {
	service service
	server  *http.Server
	port    int
	metrics metrics
}

type Config struct {
	Port int
}

const (
	readHeaderTimeout = 5 * time.Second
	gracefulTimeout   = 10 * time.Second
)

func New(port int, service service, metrics metrics) *Server {
	r := chi.NewRouter()

	s := Server{
		service: service,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		port:    port,
		metrics: metrics,
	}

	r.Route("/api/v1/xr", func(r chi.Router) {
		r.Get("/", s.readExchangeRate)
	})

	r.Get("/xr/metrics", promhttp.Handler().ServeHTTP)

	return &s
}

func (s *Server) Run(ctx context.Context) error {
	logrus.Infof("starting xr-server on port %d", s.port)

	go func() {
		<-ctx.Done()
		logrus.Infof("shutting down xr-server on port %d", s.port)

		gracefulCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		//nolint:contextcheck
		if err := s.server.Shutdown(gracefulCtx); err != nil {
			logrus.Warnf("error graceful shutdown xr-server: %v", err)

			return
		}
	}()

	if err := s.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		logrus.Errorf("error listening on port %d: %v", s.port, err)
	}

	return nil
}

func (s *Server) errorResponse(w http.ResponseWriter, errorText string, err error) {
	statusCode := http.StatusInternalServerError

	if errors.Is(err, models.ErrWrongCurrency) {
		statusCode = http.StatusNotFound
	}

	errResp := fmt.Errorf("%s: %w", errorText, err).Error()
	if statusCode == http.StatusInternalServerError {
		errResp = http.StatusText(http.StatusInternalServerError)

		logrus.Warn(err.Error())
	}

	response, err := json.Marshal(errResp)
	if err != nil {
		logrus.Warnf("error marshalling response: %v", err)
	}

	w.WriteHeader(statusCode)

	if _, err := w.Write(response); err != nil {
		logrus.Warnf("error writing response: %v", err)
	}
}

func (s *Server) okResponse(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.Warnf("error encoding response: %v", err)
	}
}

func (s *Server) readExchangeRate(w http.ResponseWriter, r *http.Request) {
	request := getQueryParams(r)

	rate, err := s.service.GetRate(request)
	if err != nil {
		s.errorResponse(w, "error getting rate", fmt.Errorf("%w", err))

		return
	}

	response := models.XRResponse{Rate: rate}

	s.metrics.TrackExternalRequest(time.Now(), r.URL.Path)

	s.okResponse(w, http.StatusOK, response)
}

func getQueryParams(r *http.Request) models.XRRequest {
	queryParams := r.URL.Query()

	xr := models.XRRequest{
		FromCurrency: queryParams.Get("from"),
		ToCurrency:   queryParams.Get("to"),
	}

	if f := queryParams.Get("from"); f != "" {
		xr.FromCurrency = f
	}

	if t := queryParams.Get("to"); t != "" {
		xr.ToCurrency = t
	}

	return xr
}
