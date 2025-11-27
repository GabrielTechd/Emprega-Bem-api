package candidates

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Experience struct {
	Title       string `bson:"title" json:"title"`
	Company     string `bson:"company" json:"company"`
	StartDate   string `bson:"start_date" json:"start_date"`
	EndDate     string `bson:"end_date,omitempty" json:"end_date,omitempty"`
	IsCurrent   bool   `bson:"is_current" json:"is_current"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
}

type Candidate struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string        `bson:"name" json:"name" validate:"required"`
	Email       string        `bson:"email" json:"email" validate:"required,email"`
	Password    string        `bson:"password" json:"-"`
	Phone       string        `bson:"phone" json:"phone"`
	Resume      string        `bson:"resume,omitempty" json:"resume,omitempty"`
	Skills      []string      `bson:"skills,omitempty" json:"skills,omitempty"`
	Experiences []Experience  `bson:"experiences,omitempty" json:"experiences,omitempty"`
	Location    string        `bson:"location" json:"location"`
	LinkedIn    string        `bson:"linkedin,omitempty" json:"linkedin,omitempty"`
	GitHub      string        `bson:"github,omitempty" json:"github,omitempty"`
	Portfolio   string        `bson:"portfolio,omitempty" json:"portfolio,omitempty"`
	CreatedAt   time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at" json:"updated_at"`
}

type CandidateRepository interface {
	Create(candidate *Candidate) error
	GetByID(id string) (*Candidate, error)
	GetByEmail(email string) (*Candidate, error)
	Update(candidate *Candidate) error
	Delete(id string) error
}
