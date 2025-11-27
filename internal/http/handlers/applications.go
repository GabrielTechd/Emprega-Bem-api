package handlers

import (
	"context"
	"empregabemapi/applications"
	"empregabemapi/candidates"
	"empregabemapi/internal/middleware"
	"empregabemapi/jobs"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ApplicationsHandler struct {
	appRepo       *applications.MongoRepository
	jobRepo       *jobs.MongoRepository
	candidateRepo *candidates.MongoRepository
}

func NewApplicationsHandler(appRepo *applications.MongoRepository, jobRepo *jobs.MongoRepository, candidateRepo *candidates.MongoRepository) *ApplicationsHandler {
	return &ApplicationsHandler{
		appRepo:       appRepo,
		jobRepo:       jobRepo,
		candidateRepo: candidateRepo,
	}
}

type ApplyRequest struct {
	JobID   string `json:"job_id"`
	Message string `json:"message,omitempty"`
}

func (h *ApplicationsHandler) Apply(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)

	var req ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	if req.JobID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "job_id é obrigatório",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se a vaga existe e está ativa
	job, err := h.jobRepo.GetByID(ctx, req.JobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	if !job.IsActive {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Esta vaga não está mais ativa",
		})
		return
	}

	// Verifica se já aplicou
	exists, err := h.appRepo.ExistsByJobAndCandidate(ctx, req.JobID, candidateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao verificar candidatura",
		})
		return
	}

	if exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Você já se candidatou a esta vaga",
		})
		return
	}

	// Converte IDs
	candidateObjID, _ := bson.ObjectIDFromHex(candidateID)
	jobObjID, _ := bson.ObjectIDFromHex(req.JobID)

	// Cria candidatura
	application := &applications.Application{
		CandidateID: candidateObjID,
		JobID:       jobObjID,
		CompanyID:   job.CompanyID,
		Message:     req.Message,
	}

	if err := h.appRepo.Create(ctx, application); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao criar candidatura",
		})
		return
	}

	// Incrementa contador de candidatos na vaga (síncrono para garantir execução)
	if err := h.jobRepo.IncrementApplicants(ctx, req.JobID); err != nil {
		// Log do erro mas não falha a requisição
		// O candidato já foi criado com sucesso
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem":    "Candidatura realizada com sucesso",
		"candidatura": application,
	})
}

func (h *ApplicationsHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	apps, err := h.appRepo.GetByCandidateID(ctx, candidateID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao buscar candidaturas",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"candidaturas": apps,
	})
}

func (h *ApplicationsHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	candidateID := r.Context().Value(middleware.UserIDKey).(string)
	appID := r.URL.Path[len("/candidate/applications/"):]

	if appID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID da candidatura não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca candidatura
	app, err := h.appRepo.GetByID(ctx, appID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Candidatura não encontrada",
		})
		return
	}

	// Verifica se é do candidato
	if app.CandidateID.Hex() != candidateID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Você não pode cancelar esta candidatura",
		})
		return
	}

	// Não permite cancelar se o status foi alterado pela empresa
	if app.Status != "pending" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Não é possível cancelar candidatura com status alterado pela empresa",
		})
		return
	}

	if err := h.appRepo.Delete(ctx, appID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao cancelar candidatura",
		})
		return
	}

	// Decrementa contador de candidatos na vaga (síncrono para garantir execução)
	if err := h.jobRepo.DecrementApplicants(ctx, app.JobID.Hex()); err != nil {
		// Log do erro mas não falha a requisição
		// A candidatura já foi deletada com sucesso
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Candidatura cancelada com sucesso",
	})
}

func (h *ApplicationsHandler) ListJobApplicants(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)
	jobID := r.URL.Path[len("/company/jobs/"):]

	// Remove "/applicants" do final
	if len(jobID) > 11 {
		jobID = jobID[:len(jobID)-11]
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
			"erro": "Você não tem permissão para ver candidatos desta vaga",
		})
		return
	}

	// Busca candidaturas
	apps, err := h.appRepo.GetByJobID(ctx, jobID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao buscar candidatos",
		})
		return
	}

	// Busca dados dos candidatos
	type ApplicantData struct {
		Application *applications.Application `json:"application"`
		Candidate   *candidates.Candidate     `json:"candidate"`
	}

	var applicants []ApplicantData
	for _, app := range apps {
		candidate, err := h.candidateRepo.GetByID(ctx, app.CandidateID.Hex())
		if err == nil {
			applicants = append(applicants, ApplicantData{
				Application: app,
				Candidate:   candidate,
			})
		}
	}

	// Sincroniza o contador de candidatos com o número real
	realCount := len(applicants)
	if realCount != job.Applicants {
		go h.jobRepo.SetApplicantsCount(context.Background(), jobID, realCount)
	}

	// Retorna com o contador correto
	job.Applicants = realCount

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vaga":       job,
		"candidatos": applicants,
	})
}

// UpdateApplicationStatus atualiza o status de uma candidatura (apenas empresa dona da vaga)
func (h *ApplicationsHandler) UpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	companyID := r.Context().Value(middleware.UserIDKey).(string)

	// Extrair application_id da URL: /company/applications/{id}/status
	// URL: /company/applications/6928b8e6a8d2e46c1b1348fd/status
	path := r.URL.Path

	// Remove o prefixo /company/applications/
	prefix := "/company/applications/"
	if !strings.HasPrefix(path, prefix) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "URL inválida"})
		return
	}

	// Remove o prefixo e o sufixo /status
	path = strings.TrimPrefix(path, prefix)
	path = strings.TrimSuffix(path, "/status")

	applicationID := path

	if applicationID == "" || len(applicationID) != 24 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID da candidatura inválido"})
		return
	}

	type UpdateStatusRequest struct {
		Status string `json:"status"`
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Dados inválidos"})
		return
	}

	// Validar status
	validStatuses := map[string]bool{
		"pending":     true,
		"viewed":      true,
		"in_review":   true,
		"shortlisted": true,
		"interview":   true,
		"rejected":    true,
		"accepted":    true,
	}

	if !validStatuses[req.Status] {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Status inválido"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Buscar candidatura
	app, err := h.appRepo.GetByID(ctx, applicationID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Candidatura não encontrada"})
		return
	}

	// Verificar se a vaga pertence à empresa
	if app.CompanyID.Hex() != companyID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Você não tem permissão para atualizar esta candidatura"})
		return
	}

	// Atualizar status
	app.Status = req.Status
	app.UpdatedAt = time.Now()

	// Se está marcando como viewed pela primeira vez
	if req.Status == "viewed" && app.ViewedAt == nil {
		now := time.Now()
		app.ViewedAt = &now
	}

	if err := h.appRepo.Update(ctx, app); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao atualizar status"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Status atualizado com sucesso"})
}
