package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "ACTIVE"
	StatusInactive  SubscriptionStatus = "INACTIVE"
	StatusCancelled SubscriptionStatus = "CANCELLED"
	StatusExpired   SubscriptionStatus = "EXPIRED"
)

type Subscription struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id" validate:"required"`
	PlanID    primitive.ObjectID `json:"plan_id" bson:"plan_id" validate:"required"`
	Plan      *Plan              `json:"plan,omitempty" bson:"-"`
	Status    SubscriptionStatus `json:"status" bson:"status"`
	StartDate time.Time          `json:"start_date" bson:"start_date"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type CreateSubscriptionRequest struct {
	UserID string             `json:"user_id" validate:"required"`
	PlanID primitive.ObjectID `json:"plan_id" validate:"required"`
}

type UpdateSubscriptionRequest struct {
	PlanID primitive.ObjectID `json:"plan_id" validate:"required"`
}
