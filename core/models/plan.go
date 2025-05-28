package models

import (
	// "time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Plan struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name" validate:"required"`
	Price    float64            `json:"price" bson:"price" validate:"required,min=0"`
	Features []string           `json:"features" bson:"features" validate:"required"`
	Duration string             `json:"duration" bson:"duration" validate:"required,oneof=monthly yearly"`
	//	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	//
	// UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
