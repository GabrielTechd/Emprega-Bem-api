package handlers

import (
	"context"
	"empregabemapi/jobs"
	"encoding/json"
	"net/http"
	"time"
)

type MaintenanceHandler struct {
	jobRepo *jobs.MongoRepository
}

func NewMaintenanceHandler(jobRepo *jobs.MongoRepository) *MaintenanceHandler {
	return &MaintenanceHandler{
		jobRepo: jobRepo,
	}
}

// FixJobCounters corrige os contadores de vagas que não foram inicializados
func (h *MaintenanceHandler) FixJobCounters(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Busca todas as vagas
	allJobs, err := h.jobRepo.List(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao buscar vagas",
		})
		return
	}

	fixed := 0
	for _, job := range allJobs {
		// Se applicants ou views estão zerados (ou não definidos), inicializa
		needsUpdate := false

		// Força inicialização dos campos
		if job.Applicants == 0 {
			needsUpdate = true
		}
		if job.Views == 0 {
			needsUpdate = true
		}

		if needsUpdate {
			// Define valores padrão
			if job.Applicants == 0 {
				job.Applicants = 0
			}
			if job.Views == 0 {
				job.Views = 0
			}

			// Atualiza no banco
			if err := h.jobRepo.Update(ctx, job); err == nil {
				fixed++
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mensagem":         "Contadores corrigidos",
		"vagas_corrigidas": fixed,
		"total_vagas":      len(allJobs),
	})
}
