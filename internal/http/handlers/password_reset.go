package handlers

import (
	"context"
	"empregabemapi/candidates"
	"empregabemapi/companies"
	"empregabemapi/internal/auth"
	"empregabemapi/internal/models"
	"empregabemapi/internal/repository"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PasswordResetHandler struct {
	resetRepo     *repository.PasswordResetRepository
	companyRepo   *companies.MongoRepository
	candidateRepo *candidates.MongoRepository
}

func NewPasswordResetHandler(
	resetRepo *repository.PasswordResetRepository,
	companyRepo *companies.MongoRepository,
	candidateRepo *candidates.MongoRepository,
) *PasswordResetHandler {
	return &PasswordResetHandler{
		resetRepo:     resetRepo,
		companyRepo:   companyRepo,
		candidateRepo: candidateRepo,
	}
}

type RequestResetRequest struct {
	Email    string `json:"email"`
	UserType string `json:"user_type"` // "company" ou "candidate"
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// RequestReset solicita reset de senha (envia "email" com token)
func (h *PasswordResetHandler) RequestReset(w http.ResponseWriter, r *http.Request) {
	var req RequestResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Validações
	if req.Email == "" || req.UserType == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Campos obrigatórios: email, user_type (company ou candidate)",
		})
		return
	}

	if req.UserType != "company" && req.UserType != "candidate" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "user_type deve ser 'company' ou 'candidate'",
		})
		return
	}

	// Sanitizar email
	req.Email = auth.SanitizeEmail(req.Email)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verificar se usuário existe (sem revelar se existe ou não por segurança)
	var userExists bool
	if req.UserType == "company" {
		_, err := h.companyRepo.GetByEmail(ctx, req.Email)
		userExists = err == nil
	} else {
		_, err := h.candidateRepo.GetByEmail(ctx, req.Email)
		userExists = err == nil
	}

	// IMPORTANTE: Sempre retornar sucesso mesmo se usuário não existir
	// Isso previne enumeration attacks (descobrir emails cadastrados)
	if userExists {
		// Gerar token de reset
		jwtToken, uniqueToken, err := auth.GeneratePasswordResetToken(req.Email, req.UserType)
		if err != nil {
			// Log do erro internamente, mas não revelar ao cliente
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"mensagem": "Se o email existir, você receberá instruções para redefinir sua senha.",
			})
			return
		}

		// Hash do token único
		tokenHash, err := auth.GenerateSecureToken(32)
		if err != nil {
			tokenHash = uniqueToken // fallback
		}

		// Salvar token no banco
		reset := &models.PasswordReset{
			Email:     req.Email,
			UserType:  req.UserType,
			Token:     jwtToken,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		if err := h.resetRepo.Create(ctx, reset); err != nil {
			// Log erro mas não revelar
		}

		// AQUI: Em produção, enviar email com link de reset
		// Link seria: https://seusite.com/reset-password?token={jwtToken}

		// Para desenvolvimento, retornar o token (REMOVER EM PRODUÇÃO!)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mensagem": "Se o email existir, você receberá instruções para redefinir sua senha.",
			// APENAS PARA DESENVOLVIMENTO - REMOVER EM PRODUÇÃO:
			"dev_token": jwtToken,
			"dev_note":  "Em produção, o token seria enviado por email",
		})
		return
	}

	// Resposta genérica (não revela se email existe)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Se o email existir, você receberá instruções para redefinir sua senha.",
	})
}

// ResetPassword redefine a senha usando o token
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Dados inválidos",
		})
		return
	}

	// Validações
	if req.Token == "" || req.NewPassword == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Campos obrigatórios: token, new_password",
		})
		return
	}

	// Validar força da nova senha
	if err := auth.ValidatePasswordStrength(req.NewPassword); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validar token JWT
	claims, err := auth.ValidatePasswordResetToken(req.Token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Token inválido ou expirado",
		})
		return
	}

	// Verificar se token não foi usado
	reset, err := h.resetRepo.GetByToken(ctx, claims.TokenHash)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Token inválido, expirado ou já utilizado",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao verificar token",
		})
		return
	}

	// Hash da nova senha
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"erro": "Erro ao processar senha",
		})
		return
	}

	// Atualizar senha no banco correto
	if claims.UserType == "company" {
		company, err := h.companyRepo.GetByEmail(ctx, claims.Email)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Usuário não encontrado",
			})
			return
		}

		company.Password = hashedPassword
		company.UpdatedAt = time.Now()

		if err := h.companyRepo.Update(ctx, company); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Erro ao atualizar senha",
			})
			return
		}
	} else {
		candidate, err := h.candidateRepo.GetByEmail(ctx, claims.Email)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Usuário não encontrado",
			})
			return
		}

		candidate.Password = hashedPassword
		candidate.UpdatedAt = time.Now()

		if err := h.candidateRepo.Update(ctx, candidate); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"erro": "Erro ao atualizar senha",
			})
			return
		}
	}

	// Marcar token como usado
	if err := h.resetRepo.MarkAsUsed(ctx, reset.ID); err != nil {
		// Log erro mas continuar (senha já foi alterada)
	}

	// Invalidar todos os outros tokens deste usuário
	h.resetRepo.InvalidateAllForUser(ctx, claims.Email, claims.UserType)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"mensagem": "Senha alterada com sucesso! Faça login com a nova senha.",
	})
}
