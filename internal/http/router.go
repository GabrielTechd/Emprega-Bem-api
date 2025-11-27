package http

import (
	"empregabemapi/applications"
	"empregabemapi/candidates"
	"empregabemapi/companies"
	"empregabemapi/internal/http/handlers"
	"empregabemapi/internal/middleware"
	"empregabemapi/internal/repository"
	"empregabemapi/jobs"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupRoutes(
	companyRepo *companies.MongoRepository,
	candidateRepo *candidates.MongoRepository,
	jobsRepo *jobs.MongoRepository,
	appsRepo *applications.MongoRepository,
	savedJobsRepo *repository.SavedJobsRepository,
	db *mongo.Database,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check
	healthHandler := handlers.NewHealthHandler()
	mux.HandleFunc("/api", healthHandler.Ping)

	// Maintenance endpoint (temporary - remove in production)
	maintenanceHandler := handlers.NewMaintenanceHandler(jobsRepo)
	mux.HandleFunc("/maintenance/fix-counters", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			maintenanceHandler.FixJobCounters(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	// Authentication handlers (com verificação cruzada de emails)
	companyAuthHandler := handlers.NewCompanyAuthHandler(companyRepo, candidateRepo)
	candidateAuthHandler := handlers.NewCandidateAuthHandler(candidateRepo, companyRepo)

	// Password reset handler
	resetRepo := repository.NewPasswordResetRepository(db)
	passwordResetHandler := handlers.NewPasswordResetHandler(resetRepo, companyRepo, candidateRepo)

	// Company authentication (public)
	mux.HandleFunc("/company/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			companyAuthHandler.Register(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/company/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			companyAuthHandler.Login(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	// Candidate authentication (public)
	mux.HandleFunc("/candidate/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			candidateAuthHandler.Register(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/candidate/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			candidateAuthHandler.Login(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	// Password reset routes (public)
	mux.HandleFunc("/auth/request-reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			passwordResetHandler.RequestReset(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/auth/reset-password", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			passwordResetHandler.ResetPassword(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	// Public jobs list (only active jobs)
	jobsHandler := handlers.NewJobsHandler(jobsRepo)
	mux.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			jobsHandler.List(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) {
		// Check for /jobs/{id}/view endpoint
		if r.Method == http.MethodPost && len(r.URL.Path) > 5 && r.URL.Path[len(r.URL.Path)-5:] == "/view" {
			jobsHandler.RegisterView(w, r)
			return
		}

		if r.Method == http.MethodGet {
			jobsHandler.GetByID(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})

	// Company profile handlers
	companyHandler := handlers.NewCompanyHandler(companyRepo)
	mux.HandleFunc("/company/me", middleware.AuthMiddleware(middleware.CompanyOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			companyHandler.GetProfile(w, r)
		} else if r.Method == http.MethodPut {
			companyHandler.UpdateProfile(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	// Application handlers (needed for company jobs applicants endpoint)
	applicationsHandler := handlers.NewApplicationsHandler(appsRepo, jobsRepo, candidateRepo)

	// Company jobs handlers
	companyJobsHandler := handlers.NewCompanyJobsHandler(jobsRepo, companyRepo)
	mux.HandleFunc("/company/jobs", middleware.AuthMiddleware(middleware.CompanyOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			companyJobsHandler.Create(w, r)
		} else if r.Method == http.MethodGet {
			companyJobsHandler.ListMine(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	mux.HandleFunc("/company/jobs/", middleware.AuthMiddleware(middleware.CompanyOnly(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check for specific actions
		if len(path) > len("/company/jobs/") {
			// Extract ID and check for action suffix
			if r.Method == http.MethodPatch {
				// Could be activate or deactivate
				if len(path) > 10 && path[len(path)-10:] == "/activate" {
					companyJobsHandler.Activate(w, r)
					return
				} else if len(path) > 12 && path[len(path)-12:] == "/deactivate" {
					companyJobsHandler.Deactivate(w, r)
					return
				}
			} else if r.Method == http.MethodGet && len(path) > 11 && path[len(path)-11:] == "/applicants" {
				applicationsHandler.ListJobApplicants(w, r)
				return
			}
		}

		// Default actions on /company/jobs/{id}
		if r.Method == http.MethodPut {
			companyJobsHandler.Update(w, r)
		} else if r.Method == http.MethodDelete {
			companyJobsHandler.Delete(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	// Company applications status update
	mux.HandleFunc("/company/applications/", middleware.AuthMiddleware(middleware.CompanyOnly(func(w http.ResponseWriter, r *http.Request) {
		// Check for /company/applications/{id}/status
		if r.Method == http.MethodPatch && len(r.URL.Path) > 7 && r.URL.Path[len(r.URL.Path)-7:] == "/status" {
			applicationsHandler.UpdateApplicationStatus(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	// Candidate profile handlers
	candidateHandler := handlers.NewCandidateHandler(candidateRepo)
	mux.HandleFunc("/candidate/me", middleware.AuthMiddleware(middleware.CandidateOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			candidateHandler.GetProfile(w, r)
		} else if r.Method == http.MethodPut {
			candidateHandler.UpdateProfile(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))
	mux.HandleFunc("/candidate/applications", middleware.AuthMiddleware(middleware.CandidateOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			applicationsHandler.Apply(w, r)
		} else if r.Method == http.MethodGet {
			applicationsHandler.ListMine(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	mux.HandleFunc("/candidate/applications/", middleware.AuthMiddleware(middleware.CandidateOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			applicationsHandler.Cancel(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	// Saved jobs handlers for candidates
	savedJobsHandler := handlers.NewSavedJobsHandler(savedJobsRepo, jobsRepo)
	mux.HandleFunc("/candidate/saved-jobs", middleware.AuthMiddleware(middleware.CandidateOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			savedJobsHandler.SaveJob(w, r)
		} else if r.Method == http.MethodGet {
			savedJobsHandler.ListSavedJobs(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	mux.HandleFunc("/candidate/saved-jobs/", middleware.AuthMiddleware(middleware.CandidateOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			savedJobsHandler.UnsaveJob(w, r)
		} else {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	return mux
}
