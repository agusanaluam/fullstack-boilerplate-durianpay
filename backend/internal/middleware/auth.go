package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func JWTAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				transport.WriteAppError(w, entity.ErrorUnauthorized("missing token"))
				return
			}
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, entity.ErrorUnauthorized("invalid signing method")
				}
				return secret, nil
			})
			if err != nil || !token.Valid {
				transport.WriteAppError(w, entity.ErrorUnauthorized("invalid token"))
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, token.Claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
