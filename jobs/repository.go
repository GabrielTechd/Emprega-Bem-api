package jobs

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection("jobs"),
	}
}

func (r *MongoRepository) Create(ctx context.Context, job *Job) error {
	job.ID = bson.NewObjectID()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	// Inicializa contadores em 0 se não foram definidos
	if job.Views == 0 {
		job.Views = 0
	}
	if job.Applicants == 0 {
		job.Applicants = 0
	}

	_, err := r.collection.InsertOne(ctx, job)
	return err
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (*Job, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var job Job
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *MongoRepository) List(ctx context.Context) ([]*Job, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []*Job
	if err = cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

type SearchFilters struct {
	Location  string
	JobType   string
	Level     string
	MinSalary float64
}

func (r *MongoRepository) Search(ctx context.Context, filters SearchFilters) ([]*Job, error) {
	filter := bson.M{}

	// Filtro de localização (case-insensitive, busca parcial)
	if filters.Location != "" {
		filter["location"] = bson.M{"$regex": filters.Location, "$options": "i"}
	}

	// Filtro de tipo de vaga (exato, case-insensitive)
	if filters.JobType != "" {
		filter["job_type"] = bson.M{"$regex": "^" + filters.JobType + "$", "$options": "i"}
	}

	// Filtro de nível (exato, case-insensitive)
	if filters.Level != "" {
		filter["level"] = bson.M{"$regex": "^" + filters.Level + "$", "$options": "i"}
	}

	// Filtro de salário mínimo
	if filters.MinSalary > 0 {
		filter["salary"] = bson.M{"$gte": filters.MinSalary}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []*Job
	if err = cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *MongoRepository) Update(ctx context.Context, job *Job) error {
	job.UpdatedAt = time.Now()
	filter := bson.M{"_id": job.ID}
	update := bson.M{"$set": job}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoRepository) Delete(ctx context.Context, id string) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// ExistsByTitleAndCompany verifica se já existe uma vaga com o mesmo título e empresa
func (r *MongoRepository) ExistsByTitleAndCompany(ctx context.Context, title, company string) (bool, error) {
	filter := bson.M{
		"title":   title,
		"company": company,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetByCompanyID retorna todas as vagas de uma empresa específica
func (r *MongoRepository) GetByCompanyID(ctx context.Context, companyID bson.ObjectID) ([]*Job, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"company_id": companyID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var jobs []*Job
	if err = cursor.All(ctx, &jobs); err != nil {
		return nil, err
	}

	return jobs, nil
}

// IncrementViews incrementa o contador de visualizações da vaga
func (r *MongoRepository) IncrementViews(ctx context.Context, jobID string) error {
	objectID, err := bson.ObjectIDFromHex(jobID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$inc": bson.M{"views": 1}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

// IncrementApplicants incrementa o contador de candidatos da vaga
func (r *MongoRepository) IncrementApplicants(ctx context.Context, jobID string) error {
	objectID, err := bson.ObjectIDFromHex(jobID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$inc": bson.M{"applicants": 1}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
} // DecrementApplicants decrementa o contador de candidatos (quando candidatura é cancelada)
func (r *MongoRepository) DecrementApplicants(ctx context.Context, jobID string) error {
	objectID, err := bson.ObjectIDFromHex(jobID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$inc": bson.M{"applicants": -1}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

// SetApplicantsCount atualiza o contador de candidatos para um valor específico
func (r *MongoRepository) SetApplicantsCount(ctx context.Context, jobID string, count int) error {
	objectID, err := bson.ObjectIDFromHex(jobID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"applicants": count}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}
