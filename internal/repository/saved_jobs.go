package repository

import (
	"context"
	"empregabemapi/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SavedJobsRepository struct {
	collection *mongo.Collection
}

func NewSavedJobsRepository(db *mongo.Database) *SavedJobsRepository {
	return &SavedJobsRepository{
		collection: db.Collection("saved_jobs"),
	}
}

// Save adiciona uma vaga aos favoritos
func (r *SavedJobsRepository) Save(ctx context.Context, candidateID, jobID string) error {
	candidateObjID, _ := bson.ObjectIDFromHex(candidateID)
	jobObjID, _ := bson.ObjectIDFromHex(jobID)

	savedJob := &models.SavedJob{
		ID:          bson.NewObjectID(),
		CandidateID: candidateObjID,
		JobID:       jobObjID,
		SavedAt:     time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, savedJob)
	return err
}

// Unsave remove uma vaga dos favoritos
func (r *SavedJobsRepository) Unsave(ctx context.Context, candidateID, jobID string) error {
	candidateObjID, _ := bson.ObjectIDFromHex(candidateID)
	jobObjID, _ := bson.ObjectIDFromHex(jobID)

	_, err := r.collection.DeleteOne(ctx, bson.M{
		"candidate_id": candidateObjID,
		"job_id":       jobObjID,
	})
	return err
}

// IsSaved verifica se uma vaga está salva
func (r *SavedJobsRepository) IsSaved(ctx context.Context, candidateID, jobID string) (bool, error) {
	candidateObjID, _ := bson.ObjectIDFromHex(candidateID)
	jobObjID, _ := bson.ObjectIDFromHex(jobID)

	count, err := r.collection.CountDocuments(ctx, bson.M{
		"candidate_id": candidateObjID,
		"job_id":       jobObjID,
	})

	return count > 0, err
}

// GetByCandidate retorna todas as vagas salvas de um candidato
func (r *SavedJobsRepository) GetByCandidate(ctx context.Context, candidateID string) ([]*models.SavedJob, error) {
	candidateObjID, _ := bson.ObjectIDFromHex(candidateID)

	cursor, err := r.collection.Find(ctx, bson.M{
		"candidate_id": candidateObjID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var savedJobs []*models.SavedJob
	if err = cursor.All(ctx, &savedJobs); err != nil {
		return nil, err
	}

	return savedJobs, nil
}
