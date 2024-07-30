package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type service interface {
	CreateWallet(ctx context.Context, wallet models.Wallet) (models.Wallet, error)
	GetWallet(ctx context.Context, walletID uuid.UUID) (models.Wallet, error)
	UpdateWallet(ctx context.Context, walletID uuid.UUID, wallet models.WalletUpdate) (models.Wallet, error)
	DeleteWallet(ctx context.Context, walletID uuid.UUID) error
}

type Server struct {
	service service
	server  *http.Server
	port    int
}

const (
	readHeaderTimeout = 5 * time.Second
	gracefulTimeout   = 10 * time.Second
)

func New(port int, service service) *Server {
	r := chi.NewRouter()
	r.Handle("/*", http.FileServer(http.Dir("./web")))

	s := Server{
		service: service,
		//nolint:exhaustivestruct
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		port: port,
	}

	r.Route("/api/v1/wallets", func(r chi.Router) {
		r.Post("/", s.createWallet)
		r.Get("/{id}", s.getWallet)
		r.Delete("/{id}", s.deleteWallet)
	})

	return &s
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

func (s *Server) errorResponse(w http.ResponseWriter, errorText string, err error) {
	statusCode := http.StatusInternalServerError

	switch {
	case errors.Is(err, models.ErrWalletNotFound):
		statusCode = http.StatusNotFound
	case errors.Is(err, models.ErrUserNotFound):
		statusCode = http.StatusNotFound
	}

	errText := fmt.Errorf("%s: %w", errorText, err).Error()
	if statusCode == http.StatusInternalServerError {
		errText = http.StatusText(http.StatusInternalServerError)

		logrus.Warn(err.Error())
	}

	response, err := json.Marshal(errText)
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

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	var (
		wallet, newWallet models.Wallet
		err               error
	)

	if err = json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		s.errorResponse(w, "error decoding request body", err)

		return
	}

	if newWallet, err = s.service.CreateWallet(r.Context(), wallet); err != nil {
		s.errorResponse(w, "error creating wallet", err)

		return
	}

	s.okResponse(w, http.StatusCreated, newWallet)
}

func (s *Server) getWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uuidTypeID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	walletInfo, err := s.service.GetWallet(r.Context(), uuidTypeID)
	if err != nil {
		s.errorResponse(w, "error reading wallet", err)

		return
	}

	s.okResponse(w, http.StatusOK, walletInfo)
}

func (s *Server) deleteWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uuidTypeID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	if err := s.service.DeleteWallet(r.Context(), uuidTypeID); err != nil {
		s.errorResponse(w, "error deleting wallet", err)

		return
	}

	s.okResponse(w, http.StatusNoContent, "wallet deleted successfully")
}
