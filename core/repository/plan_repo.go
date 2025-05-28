package repository

import (
	"context"
	"subservice/core/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PlanRepository struct {
	collection *mongo.Collection
}

func NewPlanRepository(db *mongo.Database) *PlanRepository {
	return &PlanRepository{
		collection: db.Collection("plans"),
	}
}

func (r *PlanRepository) Create(ctx context.Context, plan *models.Plan) error {
	plan.ID = primitive.NewObjectID()

	_, err := r.collection.InsertOne(ctx, plan)
	return err
}

func (r *PlanRepository) GetAll(ctx context.Context) ([]models.Plan, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []models.Plan
	if err = cursor.All(ctx, &plans); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Plan, error) {
	var plan models.Plan
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&plan)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) Update(ctx context.Context, id primitive.ObjectID, plan *models.Plan) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": plan},
	)
	return err
}

func (r *PlanRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
