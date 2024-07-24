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
	CreateWallet(ctx context.Context, wallet models.WalletInfo) error
	ReadWalletInfo(ctx context.Context, walletID uuid.UUID) (models.WalletInfo, error)
	UpdateWalletInfo(ctx context.Context, wallet models.WalletInfo) error
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
		r.Get("/", s.readWallet)
		r.Delete("/", s.deleteWallet)
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
	errResp := fmt.Errorf("%s: %w", errorText, err).Error()

	response, err := json.Marshal(errResp)
	if err != nil {
		logrus.Warnf("error marshalling response: %v", err)
	}

	w.WriteHeader(http.StatusInternalServerError)

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
	var wallet models.WalletInfo

	if err := json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		s.errorResponse(w, "error decoding request body", err)
	}

	if err := s.service.CreateWallet(r.Context(), wallet); err != nil {
		s.errorResponse(w, "error creating wallet", err)
	}

	s.okResponse(w, http.StatusOK, "wallet created successfully")
}

func (s *Server) readWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uuidTypeID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)
	}

	walletInfo, err := s.service.ReadWalletInfo(r.Context(), uuidTypeID)
	if err != nil {
		s.errorResponse(w, "error reading wallet", err)
	}

	s.okResponse(w, http.StatusOK, walletInfo)
}

func (s *Server) deleteWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uuidTypeID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)
	}

	if err := s.service.DeleteWallet(r.Context(), uuidTypeID); err != nil {
		s.errorResponse(w, "error deleting wallet", err)
	}

	s.okResponse(w, http.StatusNoContent, "wallet deleted successfully")
}
