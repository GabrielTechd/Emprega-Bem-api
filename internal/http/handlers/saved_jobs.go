package handlers

import (
	"context"
	"empregabemapi/internal/middleware"
	"empregabemapi/internal/repository"
	"empregabemapi/jobs"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SavedJobsHandler struct {
	savedJobsRepo *repository.SavedJobsRepository
	jobsRepo      *jobs.MongoRepository
}

func NewSavedJobsHandler(savedJobsRepo *repository.SavedJobsRepository, jobsRepo *jobs.MongoRepository) *SavedJobsHandler {
	return &SavedJobsHandler{
		savedJobsRepo: savedJobsRepo,
		jobsRepo:      jobsRepo,
	}
}

type SaveJobRequest struct {
	JobID string `json:"job_id"`
}

// SaveJob salva uma vaga nos favoritos
func (h *SavedJobsHandler) SaveJob(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)

	var req SaveJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Dados inválidos"})
		return
	}

	if req.JobID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "job_id é obrigatório"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verifica se a vaga existe
	_, err := h.jobsRepo.GetByID(ctx, req.JobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Vaga não encontrada"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao buscar vaga"})
		}
		return
	}

	// Verifica se já está salva
	isSaved, err := h.savedJobsRepo.IsSaved(ctx, candidateID, req.JobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao verificar vaga salva"})
		return
	}

	if isSaved {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Vaga já está nos favoritos"})
		return
	}

	// Salva a vaga
	if err := h.savedJobsRepo.Save(ctx, candidateID, req.JobID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao salvar vaga"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Vaga salva nos favoritos"})
}

// UnsaveJob remove uma vaga dos favoritos
func (h *SavedJobsHandler) UnsaveJob(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)

	// Extrair job_id da URL: /candidate/saved-jobs/{job_id}
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/candidate/saved-jobs/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "job_id é obrigatório"})
		return
	}
	jobID := parts[0]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verifica se está salva
	isSaved, err := h.savedJobsRepo.IsSaved(ctx, candidateID, jobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao verificar vaga salva"})
		return
	}

	if !isSaved {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Vaga não está nos favoritos"})
		return
	}

	// Remove a vaga dos favoritos
	if err := h.savedJobsRepo.Unsave(ctx, candidateID, jobID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao remover vaga dos favoritos"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vaga removida dos favoritos"})
}

// ListSavedJobs lista todas as vagas salvas do candidato com detalhes completos
func (h *SavedJobsHandler) ListSavedJobs(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Busca todas as vagas salvas
	savedJobs, err := h.savedJobsRepo.GetByCandidate(ctx, candidateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao buscar vagas salvas"})
		return
	}

	// Busca os detalhes completos de cada vaga
	type JobWithSavedAt struct {
		Job     *jobs.Job `json:"job"`
		SavedAt time.Time `json:"saved_at"`
	}

	var jobsWithDetails []JobWithSavedAt
	for _, savedJob := range savedJobs {
		job, err := h.jobsRepo.GetByID(ctx, savedJob.JobID.Hex())
		if err != nil {
			// Se a vaga foi deletada, pula
			continue
		}

		jobsWithDetails = append(jobsWithDetails, JobWithSavedAt{
			Job:     job,
			SavedAt: savedJob.SavedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobsWithDetails)
}
