package applications

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Application struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CandidateID bson.ObjectID `bson:"candidate_id" json:"candidate_id"`
	JobID       bson.ObjectID `bson:"job_id" json:"job_id"`
	CompanyID   bson.ObjectID `bson:"company_id" json:"company_id"`
	Status      string        `bson:"status" json:"status"` // "pending", "viewed", "in_process", "rejected", "accepted"
	Message     string        `bson:"message,omitempty" json:"message,omitempty"`
	AppliedAt   time.Time     `bson:"applied_at" json:"applied_at"`
	ViewedAt    *time.Time    `bson:"viewed_at,omitempty" json:"viewed_at,omitempty"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

type ApplicationRepository interface {
	Create(application *Application) error
	GetByID(id string) (*Application, error)
	GetByCandidateID(candidateID string) ([]*Application, error)
	GetByJobID(jobID string) ([]*Application, error)
	GetByCompanyID(companyID string) ([]*Application, error)
	ExistsByJobAndCandidate(jobID, candidateID string) (bool, error)
	Update(application *Application) error
	Delete(id string) error
}
