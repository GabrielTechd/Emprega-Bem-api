package handlers

import (
	"context"
	"empregabemapi/companies"
	"empregabemapi/internal/middleware"
	"empregabemapi/jobs"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CompanyJobsHandler struct {
	jobRepo     *jobs.MongoRepository
	companyRepo *companies.MongoRepository
}

func NewCompanyJobsHandler(jobRepo *jobs.MongoRepository, companyRepo *companies.MongoRepository) *CompanyJobsHandler {
	return &CompanyJobsHandler{
		jobRepo:     jobRepo,
		companyRepo: companyRepo,
	}
}

func (h *CompanyJobsHandler) Create(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)

	var job jobs.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Validação básica
	if job.Title == "" || job.Description == "" || job.Location == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Campos obrigatórios: title, description, location",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca nome da empresa
	company, err := h.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao buscar dados da empresa",
		})
		return
	}

	// Define company_id e company name
	companyObjID, _ := bson.ObjectIDFromHex(companyID)
	job.CompanyID = companyObjID
	job.Company = company.Name
	job.IsActive = true

	if err := h.jobRepo.Create(ctx, &job); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao criar vaga",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem": "Vaga criada com sucesso",
		"vaga":     job,
	})
}

func (h *CompanyJobsHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	companyObjID, _ := bson.ObjectIDFromHex(companyID)
	jobs, err := h.jobRepo.GetByCompanyID(ctx, companyObjID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao buscar vagas",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vagas": jobs,
	})
}

func (h *CompanyJobsHandler) Update(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)
	jobID := r.URL.Path[len("/company/jobs/"):]

	if jobID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID da vaga não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se a vaga existe e pertence à empresa
	existingJob, err := h.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	if existingJob.CompanyID.Hex() != companyID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Você não tem permissão para editar esta vaga",
		})
		return
	}

	// Decodifica novos dados
	var job jobs.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Mantém dados que não devem mudar
	job.ID = existingJob.ID
	job.CompanyID = existingJob.CompanyID
	job.Company = existingJob.Company
	job.CreatedAt = existingJob.CreatedAt

	if err := h.jobRepo.Update(ctx, &job); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao atualizar vaga",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem": "Vaga atualizada com sucesso",
		"vaga":     job,
	})
}

func (h *CompanyJobsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)
	jobID := r.URL.Path[len("/company/jobs/"):]

	if jobID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID da vaga não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se a vaga pertence à empresa
	job, err := h.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	if job.CompanyID.Hex() != companyID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Você não tem permissão para deletar esta vaga",
		})
		return
	}

	if err := h.jobRepo.Delete(ctx, jobID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao deletar vaga",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Vaga deletada com sucesso",
	})
}

func (h *CompanyJobsHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)
	jobID := r.URL.Path[len("/company/jobs/"):]

	// Remove "/deactivate" do final se existir
	if len(jobID) > 11 && jobID[len(jobID)-11:] == "/deactivate" {
		jobID = jobID[:len(jobID)-11]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job, err := h.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	if job.CompanyID.Hex() != companyID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Você não tem permissão para desativar esta vaga",
		})
		return
	}

	job.IsActive = false
	if err := h.jobRepo.Update(ctx, job); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao desativar vaga",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem": "Vaga desativada com sucesso",
		"vaga":     job,
	})
}

func (h *CompanyJobsHandler) Activate(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)
	jobID := r.URL.Path[len("/company/jobs/"):]

	// Remove "/activate" do final se existir
	if len(jobID) > 9 && jobID[len(jobID)-9:] == "/activate" {
		jobID = jobID[:len(jobID)-9]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job, err := h.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	if job.CompanyID.Hex() != companyID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Você não tem permissão para ativar esta vaga",
		})
		return
	}

	job.IsActive = true
	if err := h.jobRepo.Update(ctx, job); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao ativar vaga",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem": "Vaga ativada com sucesso",
		"vaga":     job,
	})
}
