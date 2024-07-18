package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	server http.Server
	port   int
}

const (
	readHeaderTimeout = 5 * time.Second
	gracefulTimeout   = 10 * time.Second
)

func New(port int) *Handler {
	r := chi.NewRouter()

	return &Handler{
		//nolint:exhaustivestruct
		server: http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		port: port,
	}
}

func (h *Handler) Run(ctx context.Context) error {
	logrus.Info("starting server on port ", h.port)

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	go func() {
		<-ctx.Done()
		logrus.Info("shutting down server")

		gracefulCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		//nolint:contextcheck
		if err := h.server.Shutdown(gracefulCtx); err != nil {
			logrus.Warnf("error graceful shutting down server: %v", err)

			return
		}
	}()

	if err := h.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		logrus.Errorf("error listening on port %d: %v", h.port, err)
	}

	return nil
}
