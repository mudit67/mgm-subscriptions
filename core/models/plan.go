package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Plan struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name" validate:"required"`
	Price    float64            `json:"price" bson:"price" validate:"required,min=0"`
	Features []string           `json:"features" bson:"features" validate:"required"`
	Duration string             `json:"duration" bson:"duration" validate:"required,oneof=monthly yearly"`
}

type PlanManager struct {
	collection *mongo.Collection
}

func NewPlanManager(db *mongo.Database) *PlanManager {
	return &PlanManager{
		collection: db.Collection("plans"),
	}
}

func (m *PlanManager) Create(ctx context.Context, plan *Plan) error {
	plan.ID = primitive.NewObjectID()
	_, err := m.collection.InsertOne(ctx, plan)
	return err
}

func (m *PlanManager) GetAll(ctx context.Context) ([]Plan, error) {
	cursor, err := m.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []Plan
	err = cursor.All(ctx, &plans)
	return plans, err
}

func (m *PlanManager) GetByID(ctx context.Context, id primitive.ObjectID) (*Plan, error) {
	var plan Plan
	err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&plan)
	return &plan, err
}

func (m *PlanManager) Update(ctx context.Context, id primitive.ObjectID, plan *Plan) error {
	_, err := m.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": plan})
	return err
}

func (m *PlanManager) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := m.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
