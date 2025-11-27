package handlers

import (
	"context"
	"empregabemapi/jobs"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type JobsHandler struct {
	repo *jobs.MongoRepository
}

func NewJobsHandler(repo *jobs.MongoRepository) *JobsHandler {
	return &JobsHandler{
		repo: repo,
	}
}

func (h *JobsHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se há filtros na query
	query := r.URL.Query()
	location := query.Get("location")
	jobType := query.Get("jobType")
	level := query.Get("level")
	minSalary := 0.0
	if salaryStr := query.Get("minSalary"); salaryStr != "" {
		if salary, err := parseFloat(salaryStr); err == nil {
			minSalary = salary
		}
	}

	// Se houver qualquer filtro, usa Search, senão usa List
	var jobsList []*jobs.Job
	var err error

	if location != "" || jobType != "" || level != "" || minSalary > 0 {
		filters := jobs.SearchFilters{
			Location:  location,
			JobType:   jobType,
			Level:     level,
			MinSalary: minSalary,
		}
		jobsList, err = h.repo.Search(ctx, filters)
	} else {
		jobsList, err = h.repo.List(ctx)
	}

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
		"vagas": jobsList,
	})
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func (h *JobsHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	if job.Title == "" || job.Description == "" || job.Company == "" || job.Location == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Campos obrigatórios: title, description, company, location",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se a vaga já existe (mesmo título e empresa)
	exists, err := h.repo.ExistsByTitleAndCompany(ctx, job.Title, job.Company)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao verificar vaga existente",
		})
		return
	}

	if exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Já existe uma vaga com este título nesta empresa",
		})
		return
	}

	// Define valores padrão para novos campos
	if !job.IsActive {
		job.IsActive = true // Vaga ativa por padrão
	}

	if err := h.repo.Create(ctx, &job); err != nil {
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

func (h *JobsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/jobs/"):]
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job, err := h.repo.GetByID(ctx, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)
}

func (h *JobsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/jobs/"):]
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca a vaga existente
	existingJob, err := h.repo.GetByID(ctx, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	// Decodifica os novos dados
	var job jobs.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Mantém o ID e CreatedAt originais
	job.ID = existingJob.ID
	job.CreatedAt = existingJob.CreatedAt

	// Validação básica
	if job.Title == "" || job.Description == "" || job.Company == "" || job.Location == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Campos obrigatórios: title, description, company, location",
		})
		return
	}

	if err := h.repo.Update(ctx, &job); err != nil {
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

func (h *JobsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/jobs/"):]
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se a vaga existe
	_, err := h.repo.GetByID(ctx, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	if err := h.repo.Delete(ctx, id); err != nil {
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

func (h *JobsHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	// Suporta tanto /jobs/{id} quanto /jobs/deactivate/{id}
	id := r.URL.Path[len("/jobs/"):]
	if len(id) > 11 && id[:11] == "deactivate/" {
		id = id[11:]
	}

	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca a vaga existente
	job, err := h.repo.GetByID(ctx, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	// Desativa a vaga
	job.IsActive = false

	if err := h.repo.Update(ctx, job); err != nil {
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

func (h *JobsHandler) Activate(w http.ResponseWriter, r *http.Request) {
	// Suporta tanto /jobs/{id} quanto /jobs/activate/{id}
	id := r.URL.Path[len("/jobs/"):]
	if len(id) > 9 && id[:9] == "activate/" {
		id = id[9:]
	}

	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Busca a vaga existente
	job, err := h.repo.GetByID(ctx, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	// Ativa a vaga
	job.IsActive = true

	if err := h.repo.Update(ctx, job); err != nil {
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

// RegisterView incrementa o contador de visualizações (endpoint separado para evitar duplicatas)
func (h *JobsHandler) RegisterView(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/jobs/"):]
	// Remove "/view" do final
	if len(id) > 5 {
		id = id[:len(id)-5]
	}

	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "ID não fornecido",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verifica se a vaga existe
	_, err := h.repo.GetByID(ctx, id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Vaga não encontrada",
		})
		return
	}

	// Incrementa views
	if err := h.repo.IncrementViews(ctx, id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao registrar visualização",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Visualização registrada",
	})
}
