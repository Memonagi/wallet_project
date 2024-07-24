package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Server struct {
	server *http.Server
	port   int
}

const (
	readHeaderTimeout = 5 * time.Second
	gracefulTimeout   = 10 * time.Second
)

func New(port int) *Server {
	r := chi.NewRouter()

	return &Server{
		//nolint:exhaustivestruct
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		port: port,
	}
}

func (s *Server) Run(ctx context.Context) error {
	logrus.Info("starting server on port ", s.port)

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	go func() {
		<-ctx.Done()
		logrus.Info("shutting down server on port ", s.port)

		gracefulCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		//nolint:contextcheck
		if err := s.server.Shutdown(gracefulCtx); err != nil {
			logrus.Warnf("error graceful shutting down server: %v", err)

			return
		}
	}()

	if err := s.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		logrus.Errorf("error listening on port %d: %v", s.port, err)
	}

	return nil
}
