package middleware

import (
	"context"
	"empregabemapi/internal/auth"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserTypeKey contextKey = "user_type"
)

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Headers de segurança
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Token não fornecido",
			})
			return
		}

		// Validar tamanho do header para evitar ataques
		if len(authHeader) > 1024 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Header de autorização muito grande",
			})
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Formato de token inválido",
			})
			return
		}

		// Validar que não há espaços extras no token
		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Token vazio",
			})
			return
		}

		claims, err := auth.ValidateToken(tokenStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Token inválido ou expirado",
			})
			return
		}

		// Adiciona informações do usuário no context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.ID)
		ctx = context.WithValue(ctx, UserTypeKey, claims.Type)

		next(w, r.WithContext(ctx))
	}
}

// CompanyOnly permite apenas empresas
func CompanyOnly(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userType := r.Context().Value(UserTypeKey).(string)
		if userType != "company" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Acesso restrito a empresas",
			})
			return
		}
		next(w, r)
	})
}

// CandidateOnly permite apenas candidatos
func CandidateOnly(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userType := r.Context().Value(UserTypeKey).(string)
		if userType != "candidate" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Acesso restrito a candidatos",
			})
			return
		}
		next(w, r)
	})
}
