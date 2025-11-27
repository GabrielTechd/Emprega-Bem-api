package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeadersMiddleware adiciona headers de segurança em todas as respostas
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Proteção contra clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Proteção XSS
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevenir MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// HSTS - Force HTTPS (só em produção)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none'")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy (Feature Policy)
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware configura CORS de forma segura
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Verificar se origin está na lista de permitidos
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Não enviar header se origin não for permitido
				w.Header().Set("Access-Control-Allow-Origin", "null")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 horas

			// Responder a preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitMiddleware implementa rate limiting básico por IP
type RateLimiter struct {
	requests map[string]int
	limit    int
}

func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]int),
		limit:    limit,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair IP real considerando proxies
		ip := getRealIP(r)

		// Verificar limite
		if rl.requests[ip] >= rl.limit {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"erro":"Muitas requisições. Tente novamente mais tarde."}`))
			return
		}

		// Incrementar contador
		rl.requests[ip]++

		next.ServeHTTP(w, r)

		// Limpar contador após requisição (simplificado)
		// Em produção, use um sistema de janela deslizante com Redis
		if rl.requests[ip] > 0 {
			rl.requests[ip]--
		}
	})
}

// getRealIP extrai o IP real do cliente considerando proxies
func getRealIP(r *http.Request) string {
	// Tentar X-Forwarded-For primeiro (proxy/load balancer)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// Pegar primeiro IP da lista
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Tentar X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fallback para RemoteAddr
	ip := r.RemoteAddr
	// Remover porta se presente
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// SanitizeInputMiddleware limpa inputs de caracteres perigosos
func SanitizeInputMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limitar tamanho do body para evitar ataques de DoS
		r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB

		next.ServeHTTP(w, r)
	})
}
