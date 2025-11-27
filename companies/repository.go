package companies

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
		collection: db.Collection("companies"),
	}
}

func (r *MongoRepository) Create(ctx context.Context, company *Company) error {
	company.ID = bson.NewObjectID()
	company.CreatedAt = time.Now()
	company.UpdatedAt = time.Now()
	company.VerificationStatus = "pending"

	_, err := r.collection.InsertOne(ctx, company)
	return err
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (*Company, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var company Company
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&company)
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (r *MongoRepository) GetByEmail(ctx context.Context, email string) (*Company, error) {
	var company Company
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&company)
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (r *MongoRepository) GetByCNPJ(ctx context.Context, cnpj string) (*Company, error) {
	var company Company
	err := r.collection.FindOne(ctx, bson.M{"cnpj": cnpj}).Decode(&company)
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (r *MongoRepository) Update(ctx context.Context, company *Company) error {
	company.UpdatedAt = time.Now()
	filter := bson.M{"_id": company.ID}
	update := bson.M{"$set": company}

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
