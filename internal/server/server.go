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
	CreateWallet(ctx context.Context, wallet models.Wallet, userID uuid.UUID) (models.Wallet, error)
	GetWallet(ctx context.Context, walletID, userID uuid.UUID) (models.Wallet, error)
	UpdateWallet(ctx context.Context, walletID, userID uuid.UUID, wallet models.WalletUpdate) (models.Wallet, error)
	DeleteWallet(ctx context.Context, walletID, userID uuid.UUID) error
	GetWallets(ctx context.Context, request models.GetWalletsRequest, userID uuid.UUID) ([]models.Wallet, error)
	Deposit(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error
	WithdrawMoney(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error
	Transfer(ctx context.Context, userID uuid.UUID, transaction models.Transaction) error
	GetTransactions(ctx context.Context, request models.GetWalletsRequest, walletID uuid.UUID,
		userID uuid.UUID) ([]models.Transaction, error)
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
		r.Put("/{id}/deposit", s.deposit)
		r.Put("/{id}/withdraw", s.withdrawMoney)
		r.Put("/{id}/transfer", s.transfer)
		r.Get("/{id}/transactions", s.getTransactions)
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

func getStatusCode(err error) int {
	switch {
	case errors.Is(err, models.ErrWalletNotFound) || errors.Is(err, models.ErrUserNotFound) ||
		errors.Is(err, models.ErrWrongUserID) || errors.Is(err, models.ErrEmptyID):
		return http.StatusNotFound
	case errors.Is(err, models.ErrInvalidToken):
		return http.StatusUnauthorized
	case errors.Is(err, models.ErrWrongMoney) || errors.Is(err, models.ErrWrongCurrency) ||
		errors.Is(err, models.ErrInsufficientFunds):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (s *Server) errorResponse(w http.ResponseWriter, errorText string, err error) {
	statusCode := getStatusCode(err)

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

	if _, err = w.Write(response); err != nil {
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

	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	if newWallet, err = s.service.CreateWallet(ctx, wallet, userInfo.UserID); err != nil {
		s.errorResponse(w, "error creating wallet", err)

		return
	}

	s.okResponse(w, http.StatusCreated, newWallet)
}

func (s *Server) getWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	walletID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	walletInfo, err := s.service.GetWallet(ctx, walletID, userInfo.UserID)
	if err != nil {
		s.errorResponse(w, "error reading wallet", err)

		return
	}

	s.okResponse(w, http.StatusOK, walletInfo)
}

func (s *Server) updateWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	walletID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	var wallet models.WalletUpdate

	if err = json.NewDecoder(r.Body).Decode(&wallet); err != nil {
		s.errorResponse(w, "error decoding request body", err)

		return
	}

	updatedWallet, err := s.service.UpdateWallet(ctx, walletID, userInfo.UserID, wallet)
	if err != nil {
		s.errorResponse(w, "error updating wallet", err)

		return
	}

	s.okResponse(w, http.StatusOK, updatedWallet)
}

func (s *Server) deleteWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	walletID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	if err = s.service.DeleteWallet(ctx, walletID, userInfo.UserID); err != nil {
		s.errorResponse(w, "error deleting wallet", err)

		return
	}

	s.okResponse(w, http.StatusOK, "wallet deleted successfully")
}

func (s *Server) getWallets(w http.ResponseWriter, r *http.Request) {
	request := parseGetRequest(r)
	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	wallets, err := s.service.GetWallets(ctx, request, userInfo.UserID)
	if err != nil {
		s.errorResponse(w, "error getting wallets", err)

		return
	}

	s.okResponse(w, http.StatusOK, wallets)
}

func parseGetRequest(r *http.Request) models.GetWalletsRequest {
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

func (s *Server) deposit(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction

	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		s.errorResponse(w, "error decoding request body", err)

		return
	}

	if err := s.service.Deposit(ctx, userInfo.UserID, transaction); err != nil {
		s.errorResponse(w, "deposit transaction failed", err)

		return
	}

	s.okResponse(w, http.StatusOK, "successful transaction")
}

func (s *Server) withdrawMoney(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction

	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		s.errorResponse(w, "error decoding request body", err)

		return
	}

	if err := s.service.WithdrawMoney(ctx, userInfo.UserID, transaction); err != nil {
		s.errorResponse(w, "withdraw transaction failed", err)

		return
	}

	s.okResponse(w, http.StatusOK, "successful transaction")
}

func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction

	ctx := r.Context()
	userInfo := s.getFromContext(ctx)

	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		s.errorResponse(w, "error decoding request body", err)

		return
	}

	if err := s.service.Transfer(ctx, userInfo.UserID, transaction); err != nil {
		s.errorResponse(w, "transfer transaction failed", err)

		return
	}

	s.okResponse(w, http.StatusOK, "successful transaction")
}

func (s *Server) getTransactions(w http.ResponseWriter, r *http.Request) {
	request := parseGetRequest(r)
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	walletID, err := uuid.Parse(id)
	if err != nil {
		s.errorResponse(w, "error parsing uuid", err)

		return
	}

	userInfo := s.getFromContext(ctx)

	transactions, err := s.service.GetTransactions(ctx, request, walletID, userInfo.UserID)
	if err != nil {
		s.errorResponse(w, "error getting transactions", err)

		return
	}

	s.okResponse(w, http.StatusOK, transactions)
}
