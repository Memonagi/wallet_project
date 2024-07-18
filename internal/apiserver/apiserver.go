package apiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type Apiserver struct {
	server http.Server
	port   int
}

const (
	readHeaderTimeout = 5 * time.Second
	gracefulTimeout   = 10 * time.Second
)

func New(port int) *Apiserver {
	r := chi.NewRouter()

	return &Apiserver{
		//nolint:exhaustivestruct
		server: http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		port: port,
	}
}

func (a *Apiserver) Run(ctx context.Context) error {
	logrus.Info("starting server on port ", a.port)

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	go func() {
		<-ctx.Done()
		logrus.Info("shutting down server on port ", a.port)

		gracefulCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		//nolint:contextcheck
		if err := a.server.Shutdown(gracefulCtx); err != nil {
			logrus.Warnf("error graceful shutting down server: %v", err)

			return
		}
	}()

	if err := a.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		logrus.Errorf("error listening on port %d: %v", a.port, err)
	}

	return nil
}
