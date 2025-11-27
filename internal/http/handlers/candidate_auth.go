package handlers

import (
	"context"
	"empregabemapi/candidates"
	"empregabemapi/companies"
	"empregabemapi/internal/auth"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CandidateAuthHandler struct {
	repo        *candidates.MongoRepository
	companyRepo *companies.MongoRepository
}

func NewCandidateAuthHandler(repo *candidates.MongoRepository, companyRepo *companies.MongoRepository) *CandidateAuthHandler {
	return &CandidateAuthHandler{
		repo:        repo,
		companyRepo: companyRepo,
	}
}

type RegisterCandidateRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Location string `json:"location"`
}

func (h *CandidateAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Validações
	if req.Name == "" || req.Email == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Campos obrigatórios: name, email, password",
		})
		return
	}

	// Validar força da senha
	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": err.Error(),
		})
		return
	}

	// Sanitizar email
	req.Email = auth.SanitizeEmail(req.Email)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se email já existe em candidatos
	_, err := h.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Email já cadastrado",
		})
		return
	}

	// Verifica se email já existe em empresas (verificação cruzada)
	_, err = h.companyRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Email já cadastrado como empresa. Use outro email.",
		})
		return
	}

	// Criptografa senha
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao processar senha",
		})
		return
	}

	candidate := &candidates.Candidate{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Phone:    req.Phone,
		Location: req.Location,
	}

	if err := h.repo.Create(ctx, candidate); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao criar candidato",
		})
		return
	}

	// Gera token
	token, err := auth.GenerateToken(candidate.ID.Hex(), "candidate")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao gerar token",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem":  "Candidato criado com sucesso",
		"token":     token,
		"candidato": candidate,
	})
}

func (h *CandidateAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Sanitizar email
	req.Email = auth.SanitizeEmail(req.Email)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	candidate, err := h.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Email ou senha incorretos",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao buscar candidato",
		})
		return
	}

	// Verifica senha
	if !auth.CheckPasswordHash(req.Password, candidate.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Email ou senha incorretos",
		})
		return
	}

	// Gera token
	token, err := auth.GenerateToken(candidate.ID.Hex(), "candidate")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao gerar token",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":     token,
		"candidato": candidate,
	})
}
