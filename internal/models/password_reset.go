package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// PasswordReset armazena tokens de reset de senha
type PasswordReset struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email" json:"email"`
	UserType  string        `bson:"user_type" json:"user_type"` // "company" ou "candidate"
	Token     string        `bson:"token" json:"token"`         // JWT token
	TokenHash string        `bson:"token_hash" json:"-"`        // Hash do token único
	Used      bool          `bson:"used" json:"used"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time     `bson:"expires_at" json:"expires_at"`
	UsedAt    *time.Time    `bson:"used_at,omitempty" json:"used_at,omitempty"`
}
