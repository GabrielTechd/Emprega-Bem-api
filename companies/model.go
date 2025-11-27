package companies

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Company struct {
	ID                 bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name               string        `bson:"name" json:"name" validate:"required"`
	LegalName          string        `bson:"legal_name" json:"legal_name" validate:"required"`
	CNPJ               string        `bson:"cnpj" json:"cnpj" validate:"required"`
	Email              string        `bson:"email" json:"email" validate:"required,email"`
	Password           string        `bson:"password" json:"-"`
	Phone              string        `bson:"phone" json:"phone"`
	Website            string        `bson:"website,omitempty" json:"website,omitempty"`
	Logo               string        `bson:"logo,omitempty" json:"logo,omitempty"`
	About              string        `bson:"about,omitempty" json:"about,omitempty"`
	EmployeeCount      string        `bson:"employee_count,omitempty" json:"employee_count,omitempty"` // "1-10", "11-50", "51-200", "201-500", "500+"
	Location           string        `bson:"location" json:"location"`
	Sector             string        `bson:"sector,omitempty" json:"sector,omitempty"`
	VerificationStatus string        `bson:"verification_status" json:"verification_status"` // "pending", "verified", "rejected"
	CreatedAt          time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `bson:"updated_at" json:"updated_at"`
}

type CompanyRepository interface {
	Create(company *Company) error
	GetByID(id string) (*Company, error)
	GetByEmail(email string) (*Company, error)
	GetByCNPJ(cnpj string) (*Company, error)
	Update(company *Company) error
	Delete(id string) error
}
