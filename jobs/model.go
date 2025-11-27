package jobs

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Job struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CompanyID   bson.ObjectID `bson:"company_id" json:"company_id"` // NOVO: ID da empresa dona
	Title       string        `bson:"title" json:"title" validate:"required"`
	Description string        `bson:"description" json:"description" validate:"required"`
	Company     string        `bson:"company" json:"company" validate:"required"` // Nome da empresa (para exibição)
	Location    string        `bson:"location" json:"location" validate:"required"`
	Salary      float64       `bson:"salary,omitempty" json:"salary,omitempty"`

	// 1 — ESSENCIAL DO MVP
	JobType string `bson:"job_type,omitempty" json:"job_type,omitempty"` // remoto | presencial | híbrido
	Level   string `bson:"level,omitempty" json:"level,omitempty"`       // junior | pleno | senior

	// 2 — COMPLETA A VAGA E MELHORA BUSCA
	Requirements []string `bson:"requirements,omitempty" json:"requirements,omitempty"` // ["Go", "Docker"]
	Benefits     []string `bson:"benefits,omitempty" json:"benefits,omitempty"`         // ["VR", "Plano de saúde"]

	// 3 — STATUS DA VAGA
	IsActive bool `bson:"is_active" json:"is_active"` // vaga ativa ou encerrada

	// 4 — METADADOS
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`

	// 5 — COISAS QUE FAZEM SENTIDO EM UM SITE PROFISSIONAL
	Views      int `bson:"views" json:"views"`           // contagem de visualizações
	Applicants int `bson:"applicants" json:"applicants"` // número de candidatos aplicados
	Priority   int `bson:"priority" json:"priority"`     // 0 normal, 1 destaque
}
type JobRepository interface {
	Create(job *Job) error
	GetByID(id string) (*Job, error)
	List() ([]*Job, error)
	Update(job *Job) error
	Delete(id string) error
}
