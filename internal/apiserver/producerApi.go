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

type ProducerAPI struct {
	server http.Server
	port   int
}

func NewProducerAPI(port int) *ProducerAPI {
	r := chi.NewRouter()

	return &ProducerAPI{
		server: http.Server{
			Addr:        fmt.Sprintf(":%d", port),
			Handler:     r,
			ReadTimeout: readHeaderTimeout,
		},
		port: port,
	}
}

func (api *ProducerAPI) RunProducerAPI(ctx context.Context) error {
	logrus.Info("Starting producer server on port: ", api.port)

	t := time.NewTicker(time.Minute)
	defer t.Stop()

	go func() {
		<-ctx.Done()

		logrus.Info("Shutting down producer server on port: ", api.port)

		gfCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		//nolint:contextcheck
		if err := api.server.Shutdown(gfCtx); err != nil {
			logrus.Errorf("Error shutting down producer server: %v", err)

			return
		}
	}()

	if err := api.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		logrus.Errorf("Error starting producer server on port %d: %v", api.port, err)
	}

	return nil
}
