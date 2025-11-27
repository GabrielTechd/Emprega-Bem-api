package repository

import (
	"context"
	"empregabemapi/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PasswordResetRepository struct {
	collection *mongo.Collection
}

func NewPasswordResetRepository(db *mongo.Database) *PasswordResetRepository {
	return &PasswordResetRepository{
		collection: db.Collection("password_resets"),
	}
}

// Create salva um novo token de reset
func (r *PasswordResetRepository) Create(ctx context.Context, reset *models.PasswordReset) error {
	reset.ID = bson.NewObjectID()
	reset.CreatedAt = time.Now()
	reset.Used = false

	_, err := r.collection.InsertOne(ctx, reset)
	return err
}

// GetByToken busca um reset por token hash
func (r *PasswordResetRepository) GetByToken(ctx context.Context, tokenHash string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	err := r.collection.FindOne(ctx, bson.M{
		"token_hash": tokenHash,
		"used":       false,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&reset)

	if err != nil {
		return nil, err
	}

	return &reset, nil
}

// MarkAsUsed marca um token como usado
func (r *PasswordResetRepository) MarkAsUsed(ctx context.Context, id bson.ObjectID) error {
	now := time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"used":    true,
				"used_at": now,
			},
		},
	)
	return err
}

// DeleteExpired remove tokens expirados (limpeza)
func (r *PasswordResetRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	})
	return err
}

// InvalidateAllForUser invalida todos os tokens de um usuário (por segurança)
func (r *PasswordResetRepository) InvalidateAllForUser(ctx context.Context, email, userType string) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{
			"email":     email,
			"user_type": userType,
			"used":      false,
		},
		bson.M{
			"$set": bson.M{
				"used": true,
			},
		},
	)
	return err
}
