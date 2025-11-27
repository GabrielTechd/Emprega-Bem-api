package applications

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
		collection: db.Collection("applications"),
	}
}

func (r *MongoRepository) Create(ctx context.Context, application *Application) error {
	application.ID = bson.NewObjectID()
	application.AppliedAt = time.Now()
	application.UpdatedAt = time.Now()
	application.Status = "pending"

	_, err := r.collection.InsertOne(ctx, application)
	return err
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (*Application, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var application Application
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&application)
	if err != nil {
		return nil, err
	}

	return &application, nil
}

func (r *MongoRepository) GetByCandidateID(ctx context.Context, candidateID string) ([]*Application, error) {
	objectID, err := bson.ObjectIDFromHex(candidateID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"candidate_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var applications []*Application
	if err = cursor.All(ctx, &applications); err != nil {
		return nil, err
	}

	return applications, nil
}

func (r *MongoRepository) GetByJobID(ctx context.Context, jobID string) ([]*Application, error) {
	objectID, err := bson.ObjectIDFromHex(jobID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"job_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var applications []*Application
	if err = cursor.All(ctx, &applications); err != nil {
		return nil, err
	}

	return applications, nil
}

func (r *MongoRepository) GetByCompanyID(ctx context.Context, companyID string) ([]*Application, error) {
	objectID, err := bson.ObjectIDFromHex(companyID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{"company_id": objectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var applications []*Application
	if err = cursor.All(ctx, &applications); err != nil {
		return nil, err
	}

	return applications, nil
}

func (r *MongoRepository) ExistsByJobAndCandidate(ctx context.Context, jobID, candidateID string) (bool, error) {
	jobObjID, err := bson.ObjectIDFromHex(jobID)
	if err != nil {
		return false, err
	}

	candidateObjID, err := bson.ObjectIDFromHex(candidateID)
	if err != nil {
		return false, err
	}

	filter := bson.M{
		"job_id":       jobObjID,
		"candidate_id": candidateObjID,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *MongoRepository) Update(ctx context.Context, application *Application) error {
	application.UpdatedAt = time.Now()
	filter := bson.M{"_id": application.ID}
	update := bson.M{"$set": application}

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
