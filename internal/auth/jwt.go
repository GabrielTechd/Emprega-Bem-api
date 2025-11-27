package auth

import (
	"crypto/subtle"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret []byte

const (
	// Aumentar custo do bcrypt para maior segurança (padrão é 10, usaremos 12)
	bcryptCost = 12
	// Requisitos mínimos de senha
	minPasswordLength = 8
)

// Initialize configura o segredo JWT
func Initialize(secret string) {
	if len(secret) < 32 {
		panic("JWT_SECRET deve ter no mínimo 32 caracteres")
	}
	jwtSecret = []byte(secret)
}

type Claims struct {
	ID   string `json:"id"`
	Type string `json:"type"` // "company" ou "candidate"
	jwt.RegisteredClaims
}

// ValidatePasswordStrength verifica se a senha atende requisitos mínimos
func ValidatePasswordStrength(password string) error {
	if len(password) < minPasswordLength {
		return errors.New("senha deve ter no mínimo 8 caracteres")
	}

	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	)

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("senha deve conter letras maiúsculas, minúsculas, números e caracteres especiais")
	}

	return nil
}

// SanitizeEmail normaliza email para evitar duplicatas
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// HashPassword criptografa a senha com bcrypt cost 12
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	return string(bytes), err
}

// CheckPasswordHash verifica se a senha está correta (protegido contra timing attacks)
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// Usar subtle.ConstantTimeCompare para evitar timing attacks
	return subtle.ConstantTimeCompare([]byte{boolToByte(err == nil)}, []byte{1}) == 1
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// GenerateToken cria um token JWT com claims de segurança
func GenerateToken(id, userType string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT secret não inicializado")
	}

	claims := Claims{
		ID:   id,
		Type: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "empregabem-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken valida e retorna os claims do token com verificações de segurança
func ValidateToken(tokenString string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("JWT secret não inicializado")
	}

	// Validar tamanho do token para evitar ataques de DoS
	if len(tokenString) > 1024 {
		return nil, errors.New("token muito grande")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar algoritmo para evitar ataques de confusão de algoritmo
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("algoritmo de assinatura inválido")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Verificar issuer
		if claims.Issuer != "empregabem-api" {
			return nil, errors.New("issuer inválido")
		}
		return claims, nil
	}

	return nil, errors.New("token inválido")
}
