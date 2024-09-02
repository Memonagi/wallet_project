package server

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
	GetWallets(ctx context.Context, request models.GetWalletsRequest) ([]models.Wallet, error)
}

type Server struct {
	service service
	server  *http.Server
	key     *rsa.PublicKey
	port    int
}

type Config struct {
	Port int
}

const (
	readHeaderTimeout = 5 * time.Second
	gracefulTimeout   = 10 * time.Second
	DefaultLimit      = 25
)

func New(cfg Config, service service, key *rsa.PublicKey) *Server {
	r := chi.NewRouter()

	s := Server{
		service: service,
		//nolint:exhaustivestruct
		server: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           r,
			ReadHeaderTimeout: readHeaderTimeout,
		},
		key:  key,
		port: cfg.Port,
	}

	r.Route("/api/v1/wallets", func(r chi.Router) {
		r.Use(s.JWTCheck)

		r.Post("/", s.createWallet)
		r.Get("/{id}", s.getWallet)
		r.Patch("/{id}", s.updateWallet)
		r.Delete("/{id}", s.deleteWallet)
		r.Get("/", s.getWallets)
	})

	return &s
}

func (s *Server) Run(ctx context.Context) error {
	logrus.Info("starting server on port ", s.port)

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
	case errors.Is(err, models.ErrEmptyID):
		statusCode = http.StatusNotFound
	case errors.Is(err, models.ErrInvalidToken):
		statusCode = http.StatusUnauthorized
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

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
	var (
		wallet, newWallet models.Wallet
		err               error
	)

	if err = json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		s.errorResponse(w, "error decoding request body", err)

		return
	}

	userInfo := s.getFromContext(r.Context())

	wallet.UserID = userInfo.UserID

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

	userInfo := s.getFromContext(r.Context())

	walletInfo, err := s.service.GetWallet(r.Context(), uuidTypeID)
	if err != nil {
		s.errorResponse(w, "error reading wallet", err)

		return
	}

	if err = s.hasAccessToWallet(userInfo, walletInfo.UserID); err != nil {
		s.errorResponse(w, "error access denied", err)

		return
	}

	s.okResponse(w, http.StatusOK, walletInfo)
}

func (s *Server) updateWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uuidTypeID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	userInfo := s.getFromContext(r.Context())

	walletInfo, err := s.service.GetWallet(r.Context(), uuidTypeID)
	if err != nil {
		s.errorResponse(w, "error reading wallet", err)

		return
	}

	if err = s.hasAccessToWallet(userInfo, walletInfo.UserID); err != nil {
		s.errorResponse(w, "error access denied", err)

		return
	}

	var wallet models.WalletUpdate

	if err = json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		s.errorResponse(w, "error decoding request body", err)

		return
	}

	updatedWallet, err := s.service.UpdateWallet(r.Context(), uuidTypeID, wallet)
	if err != nil {
		s.errorResponse(w, "error updating wallet", err)

		return
	}

	s.okResponse(w, http.StatusOK, updatedWallet)
}

func (s *Server) deleteWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	uuidTypeID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	userInfo := s.getFromContext(r.Context())

	walletInfo, err := s.service.GetWallet(r.Context(), uuidTypeID)
	if err != nil {
		s.errorResponse(w, "error reading wallet", err)

		return
	}

	if err = s.hasAccessToWallet(userInfo, walletInfo.UserID); err != nil {
		s.errorResponse(w, "error access denied", err)

		return
	}

	if err := s.service.DeleteWallet(r.Context(), uuidTypeID); err != nil {
		s.errorResponse(w, "error deleting wallet", err)

		return
	}

	s.okResponse(w, http.StatusOK, "wallet deleted successfully")
}

func (s *Server) getWallets(w http.ResponseWriter, r *http.Request) {
	request := parseGetWalletsRequest(r)

	userInfo := s.getFromContext(r.Context())

	wallets, err := s.service.GetWallets(r.Context(), request)
	if err != nil {
		s.errorResponse(w, "error getting wallets", err)

		return
	}

	var usersWallets []models.Wallet

	for _, wallet := range wallets {
		if wallet.UserID == userInfo.UserID {
			usersWallets = append(usersWallets, wallet)
		}
	}

	s.okResponse(w, http.StatusOK, usersWallets)
}

func parseGetWalletsRequest(r *http.Request) models.GetWalletsRequest {
	queryParams := r.URL.Query()

	g := models.GetWalletsRequest{
		Sorting: queryParams.Get("sorting"),
		Filter:  queryParams.Get("filter"),
	}

	var (
		limit  int64
		offset int64
	)

	if d := queryParams.Get("descending"); d != "" {
		g.Descending, _ = strconv.ParseBool(d)
	}

	if l := queryParams.Get("limit"); l != "" {
		if limit, _ = strconv.ParseInt(l, 0, 64); limit == 0 {
			limit = DefaultLimit
		}
	}

	if o := queryParams.Get("offset"); o != "" {
		offset, _ = strconv.ParseInt(o, 0, 64)
	}

	g.Limit = int(limit)
	g.Offset = int(offset)

	return g
}
