package handlers

import (
	"context"
	"empregabemapi/candidates"
	"empregabemapi/internal/middleware"
	"encoding/json"
	"net/http"
	"time"
)

type CandidateHandler struct {
	repo *candidates.MongoRepository
}

func NewCandidateHandler(repo *candidates.MongoRepository) *CandidateHandler {
	return &CandidateHandler{
		repo: repo,
	}
}

func (h *CandidateHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	candidate, err := h.repo.GetByID(ctx, candidateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Candidato não encontrado",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(candidate)
}

func (h *CandidateHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca candidato existente
	candidate, err := h.repo.GetByID(ctx, candidateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Candidato não encontrado",
		})
		return
	}

	// Decodifica novos dados
	var updateData candidates.Candidate
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Atualiza campos permitidos
	candidate.Name = updateData.Name
	candidate.Phone = updateData.Phone
	candidate.Resume = updateData.Resume
	candidate.Skills = updateData.Skills
	candidate.Experiences = updateData.Experiences
	candidate.Location = updateData.Location
	candidate.LinkedIn = updateData.LinkedIn
	candidate.GitHub = updateData.GitHub
	candidate.Portfolio = updateData.Portfolio

	if err := h.repo.Update(ctx, candidate); err != nil {
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
		"mensagem":  "Perfil atualizado com sucesso",
		"candidato": candidate,
	})
}
