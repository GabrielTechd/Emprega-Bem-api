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

type CompanyAuthHandler struct {
	repo          *companies.MongoRepository
	candidateRepo *candidates.MongoRepository
}

func NewCompanyAuthHandler(repo *companies.MongoRepository, candidateRepo *candidates.MongoRepository) *CompanyAuthHandler {
	return &CompanyAuthHandler{
		repo:          repo,
		candidateRepo: candidateRepo,
	}
}

type RegisterCompanyRequest struct {
	Name      string `json:"name"`
	LegalName string `json:"legal_name"`
	CNPJ      string `json:"cnpj"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *CompanyAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Validações
	if req.Name == "" || req.Email == "" || req.Password == "" || req.CNPJ == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Campos obrigatórios: name, email, password, cnpj",
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

	// Validar formato de CNPJ (14 dígitos)
	if len(req.CNPJ) != 14 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "CNPJ deve conter exatamente 14 dígitos",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se email já existe em empresas
	_, err := h.repo.GetByEmail(ctx, req.Email)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Email já cadastrado",
		})
		return
	}

	// Verifica se email já existe em candidatos (verificação cruzada)
	_, err = h.candidateRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Email já cadastrado como candidato. Use outro email.",
		})
		return
	}

	// Verifica se CNPJ já existe
	_, err = h.repo.GetByCNPJ(ctx, req.CNPJ)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "CNPJ já cadastrado",
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

	company := &companies.Company{
		Name:      req.Name,
		LegalName: req.LegalName,
		CNPJ:      req.CNPJ,
		Email:     req.Email,
		Password:  hashedPassword,
		Phone:     req.Phone,
		Location:  req.Location,
	}

	if err := h.repo.Create(ctx, company); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao criar empresa",
		})
		return
	}

	// Gera token
	token, err := auth.GenerateToken(company.ID.Hex(), "company")
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
		"mensagem": "Empresa criada com sucesso",
		"token":    token,
		"empresa":  company,
	})
}

func (h *CompanyAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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

	company, err := h.repo.GetByEmail(ctx, req.Email)
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
			"erro": "Erro ao buscar empresa",
		})
		return
	}

	// Verifica senha
	if !auth.CheckPasswordHash(req.Password, company.Password) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Email ou senha incorretos",
		})
		return
	}

	// Gera token
	token, err := auth.GenerateToken(company.ID.Hex(), "company")
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
		"token":   token,
		"empresa": company,
	})
}
