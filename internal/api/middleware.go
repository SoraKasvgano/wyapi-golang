package api

import (
	"net/http"
	"strings"

	"github.com/go-chi/cors"
	"wyapi-golang/internal/config"
	"wyapi-golang/pkg/response"
)

func CORSMiddleware(cfg config.CORSConfig) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           300,
	})
	return c.Handler
}

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	if cfg == nil || !cfg.Security.RequireToken {
		return func(next http.Handler) http.Handler { return next }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" || token != cfg.Security.APIToken {
				response.Error(w, http.StatusUnauthorized, "无效的API Token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) string {
	if r == nil {
		return ""
	}

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[7:])
	}

	if token := r.Header.Get("X-API-Token"); token != "" {
		return token
	}

	if token := r.Header.Get("X-API-Key"); token != "" {
		return token
	}

	queryToken := r.URL.Query().Get("token")
	if queryToken != "" {
		return queryToken
	}
	queryToken = r.URL.Query().Get("api_token")
	if queryToken != "" {
		return queryToken
	}

	return ""
}
