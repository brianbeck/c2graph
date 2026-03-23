package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// Middleware returns an HTTP middleware that verifies Supabase JWTs.
type Middleware struct {
	jwtSecret string
}

// NewMiddleware creates a new auth middleware.
// If jwtSecret is empty, auth is disabled (development mode).
func NewMiddleware(jwtSecret string) *Middleware {
	return &Middleware{jwtSecret: jwtSecret}
}

// Handler returns the middleware handler function.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if no JWT secret configured (dev mode)
		if m.jwtSecret == "" || m.jwtSecret == "your-jwt-secret" {
			log.Warn().Msg("Auth disabled: no SUPABASE_JWT_SECRET configured")
			ctx := context.WithValue(r.Context(), UserIDKey, "dev-user")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract Bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			http.Error(w, `{"error": "invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		// Verify JWT
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.jwtSecret), nil
		})
		if err != nil {
			log.Debug().Err(err).Msg("JWT verification failed")
			http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, `{"error": "invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		// Extract user ID (sub claim)
		sub, _ := claims["sub"].(string)
		if sub == "" {
			http.Error(w, `{"error": "missing sub claim"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
