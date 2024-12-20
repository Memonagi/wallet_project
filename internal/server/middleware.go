package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	jwtclaims "github.com/Memonagi/wallet_project/internal/jwt-claims"
	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ctxKey contextKey = "ctxKey"
)

func (s *Server) jwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			s.errorResponse(w, "authorization error", models.ErrInvalidToken)

			return
		}

		headerParts := strings.Split(header, " ")

		if headerParts[0] != "Bearer" {
			s.errorResponse(w, "authorization error", models.ErrInvalidToken)

			return
		}

		token, err := jwt.ParseWithClaims(headerParts[1], &jwtclaims.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, models.ErrInvalidSigningMethod
			}

			return s.key, nil
		})
		if err != nil {
			s.errorResponse(w, "authorization error", models.ErrInvalidToken)

			return
		}

		claims, ok := token.Claims.(*jwtclaims.Claims)
		if !(ok && token.Valid) {
			s.errorResponse(w, "authorization error", models.ErrInvalidToken)

			return
		}

		if claims.ExpiresAt.Before(time.Now()) {
			s.errorResponse(w, "authorization error", models.ErrInvalidToken)

			return
		}

		userInfo := models.UserInfo{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		}

		r = r.WithContext(context.WithValue(r.Context(), ctxKey, userInfo))
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getFromContext(ctx context.Context) models.UserInfo {
	userInfo, _ := ctx.Value(ctxKey).(models.UserInfo)

	return userInfo
}

func (s *Server) metricTrack(next http.Handler) http.Handler {
	var fn http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		defer s.metrics.trackHTTPRequest(time.Now(), r)

		next.ServeHTTP(w, r)
	}

	return fn
}
