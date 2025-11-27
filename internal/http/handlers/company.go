package handlers

import (
	"context"
	"empregabemapi/companies"
	"empregabemapi/internal/middleware"
	"encoding/json"
	"net/http"
	"time"
)

type CompanyHandler struct {
	repo *companies.MongoRepository
}

func NewCompanyHandler(repo *companies.MongoRepository) *CompanyHandler {
	return &CompanyHandler{
		repo: repo,
	}
}

func (h *CompanyHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	company, err := h.repo.GetByID(ctx, companyID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Empresa não encontrada",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(company)
}

func (h *CompanyHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca empresa existente
	company, err := h.repo.GetByID(ctx, companyID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Empresa não encontrada",
		})
		return
	}

	// Decodifica novos dados
	var updateData companies.Company
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Mantém dados que não devem ser alterados
	company.Name = updateData.Name
	company.LegalName = updateData.LegalName
	company.Phone = updateData.Phone
	company.Website = updateData.Website
	company.Logo = updateData.Logo
	company.About = updateData.About
	company.EmployeeCount = updateData.EmployeeCount
	company.Location = updateData.Location
	company.Sector = updateData.Sector

	if err := h.repo.Update(ctx, company); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao atualizar perfil",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem": "Perfil atualizado com sucesso",
		"empresa":  company,
	})
}
