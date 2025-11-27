package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ResetTokenClaims contém informações do token de reset
type ResetTokenClaims struct {
	Email     string `json:"email"`
	UserType  string `json:"user_type"`  // "company" ou "candidate"
	TokenHash string `json:"token_hash"` // Hash do token para invalidação
	jwt.RegisteredClaims
}

// GeneratePasswordResetToken gera um token seguro para reset de senha
// Retorna: token JWT, token único (para email), erro
func GeneratePasswordResetToken(email, userType string) (string, string, error) {
	if len(jwtSecret) == 0 {
		return "", "", errors.New("JWT secret não inicializado")
	}

	// Gerar token único aleatório de 32 bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", err
	}
	uniqueToken := hex.EncodeToString(randomBytes)

	// Hash do token único para armazenar no JWT
	tokenHash := hashString(uniqueToken)

	claims := ResetTokenClaims{
		Email:     SanitizeEmail(email),
		UserType:  userType,
		TokenHash: tokenHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutos
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "empregabem-api-reset",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", err
	}

	return jwtToken, uniqueToken, nil
}

// ValidatePasswordResetToken valida token de reset de senha
func ValidatePasswordResetToken(tokenString string) (*ResetTokenClaims, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT secret não inicializado")
	}

	// Validar tamanho do token
	if len(tokenString) > 1024 {
		return nil, errors.New("token muito grande")
	}

	token, err := jwt.ParseWithClaims(tokenString, &ResetTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritmo de assinatura inválido")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*ResetTokenClaims); ok && token.Valid {
		// Verificar issuer específico de reset
		if claims.Issuer != "empregabem-api-reset" {
			return nil, errors.New("token não é de reset de senha")
		}

		// Verificar se não expirou
		if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("token de reset expirado")
		}

		return claims, nil
	}

	return nil, errors.New("token inválido")
}

// hashString cria um hash simples de uma string
func hashString(s string) string {
	// Usar bcrypt seria mais seguro mas mais lento
	// Para tokens temporários de 15min, um hash simples é suficiente
	hash := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		hash = append(hash, s[i]^0xAA)
	}
	return hex.EncodeToString(hash)
}

// GenerateSecureToken gera um token aleatório seguro
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
