package main

import (
	"empregabemapi/applications"
	"empregabemapi/candidates"
	"empregabemapi/companies"
	"empregabemapi/database"
	"empregabemapi/internal/auth"
	"empregabemapi/internal/config"
	"empregabemapi/internal/http"
	"empregabemapi/internal/middleware"
	"empregabemapi/internal/repository"
	"empregabemapi/jobs"
	"fmt"
	"log"
	nethttp "net/http"
)

func main() {
	cfg := config.Load()

	// Validar configurações críticas de segurança
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET não configurado no .env")
	}
	if len(cfg.JWTSecret) < 32 {
		log.Fatal("JWT_SECRET deve ter no mínimo 32 caracteres para segurança adequada")
	}
	auth.Initialize(cfg.JWTSecret)

	// Conectar ao MongoDB
	mongodb, err := database.NewMongoDB(cfg.DatabaseURL, "empregabem")
	if err != nil {
		log.Fatal("Erro ao conectar no MongoDB:", err)
	}
	defer mongodb.Close()

	// Criar repositórios
	companyRepo := companies.NewMongoRepository(mongodb.Database)
	candidateRepo := candidates.NewMongoRepository(mongodb.Database)
	jobsRepo := jobs.NewMongoRepository(mongodb.Database)
	appsRepo := applications.NewMongoRepository(mongodb.Database)
	savedJobsRepo := repository.NewSavedJobsRepository(mongodb.Database)

	// Configurar rotas (passando database para password reset)
	router := http.SetupRoutes(companyRepo, candidateRepo, jobsRepo, appsRepo, savedJobsRepo, mongodb.Database)

	// Configurar CORS - permitir frontend
	allowedOrigins := parseOrigins(cfg.CORSOrigins)
	corsRouter := middleware.CORSMiddleware(allowedOrigins)(router)

	// Aplicar middlewares de segurança
	secureRouter := middleware.SecurityHeadersMiddleware(corsRouter)
	secureRouter = middleware.SanitizeInputMiddleware(secureRouter)

	// Rate limiting (100 requisições por janela)
	rateLimiter := middleware.NewRateLimiter(100)
	secureRouter = rateLimiter.Middleware(secureRouter)

	addr := ":" + cfg.Port
	fmt.Printf("🚀 API rodando em %s\n", addr)
	fmt.Printf("🔒 Segurança habilitada: bcrypt cost 12, validação de senha forte, headers de segurança\n")
	fmt.Printf("🌐 CORS habilitado para: %v\n", allowedOrigins)

	if err := nethttp.ListenAndServe(addr, secureRouter); err != nil {
		log.Fatal("Erro ao iniciar api:", err)
	}
}

// parseOrigins converte string de origins separados por vírgula em slice
func parseOrigins(origins string) []string {
	if origins == "" {
		return []string{"http://localhost:3000"}
	}
	var result []string
	for _, origin := range splitAndTrim(origins, ",") {
		if origin != "" {
			result = append(result, origin)
		}
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for i := 0; i < len(s); {
		// Encontrar próximo separador
		j := i
		for j < len(s) && string(s[j]) != sep {
			j++
		}
		// Adicionar substring trimada
		substr := s[i:j]
		// Trim manual
		start := 0
		end := len(substr)
		for start < end && (substr[start] == ' ' || substr[start] == '\t') {
			start++
		}
		for start < end && (substr[end-1] == ' ' || substr[end-1] == '\t') {
			end--
		}
		if start < end {
			result = append(result, substr[start:end])
		}
		// Avançar além do separador
		i = j + len(sep)
		if i > len(s) {
			break
		}
	}
	return result
}
