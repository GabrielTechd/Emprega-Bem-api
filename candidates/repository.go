package candidates

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
		collection: db.Collection("candidates"),
	}
}

func (r *MongoRepository) Create(ctx context.Context, candidate *Candidate) error {
	candidate.ID = bson.NewObjectID()
	candidate.CreatedAt = time.Now()
	candidate.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, candidate)
	return err
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (*Candidate, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var candidate Candidate
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&candidate)
	if err != nil {
		return nil, err
	}

	return &candidate, nil
}

func (r *MongoRepository) GetByEmail(ctx context.Context, email string) (*Candidate, error) {
	var candidate Candidate
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&candidate)
	if err != nil {
		return nil, err
	}

	return &candidate, nil
}

func (r *MongoRepository) Update(ctx context.Context, candidate *Candidate) error {
	candidate.UpdatedAt = time.Now()
	filter := bson.M{"_id": candidate.ID}
	update := bson.M{"$set": candidate}

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
