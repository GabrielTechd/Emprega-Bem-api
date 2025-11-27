package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SavedJob representa uma vaga salva/favoritada por um candidato
type SavedJob struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CandidateID bson.ObjectID `bson:"candidate_id" json:"candidate_id"`
	JobID       bson.ObjectID `bson:"job_id" json:"job_id"`
	SavedAt     time.Time     `bson:"saved_at" json:"saved_at"`
}
