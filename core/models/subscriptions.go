package models

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

type SubscriptionManager struct {
	collection  *mongo.Collection
	planManager *PlanManager
}

func NewSubscriptionManager(db *mongo.Database, planManager *PlanManager) *SubscriptionManager {
	manager := &SubscriptionManager{
		collection:  db.Collection("subscriptions"),
		planManager: planManager,
	}
	manager.createIndexes()
	return manager
}

func (m *SubscriptionManager) createIndexes() {
	ctx := context.Background()
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: &options.IndexOptions{Unique: &[]bool{true}[0]},
	}
	m.collection.Indexes().CreateOne(ctx, indexModel)
}

func (m *SubscriptionManager) UpsertSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*Subscription, error) {
	// Get plan details
	plan, err := m.planManager.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	// Calculate expiry date
	now := time.Now()
	var expiryDate time.Time
	switch plan.Duration {
	case "monthly":
		expiryDate = now.AddDate(0, 1, 0)
	case "yearly":
		expiryDate = now.AddDate(1, 0, 0)
	default:
		return nil, errors.New("invalid plan duration")
	}

	subscription := &Subscription{
		UserID:    req.UserID,
		PlanID:    req.PlanID,
		Status:    StatusActive,
		StartDate: now,
		ExpiresAt: expiryDate,
		CreatedAt: now,
	}

	// Upsert subscription
	filter := bson.M{"user_id": req.UserID}
	update := bson.M{
		"$set": bson.M{
			"plan_id":    subscription.PlanID,
			"status":     subscription.Status,
			"start_date": subscription.StartDate,
			"expires_at": subscription.ExpiresAt,
			"created_at": subscription.CreatedAt,
		},
		"$setOnInsert": bson.M{"_id": primitive.NewObjectID()},
	}

	opts := options.Update().SetUpsert(true)
	_, err = m.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, err
	}

	subscription.Plan = plan
	log.Printf("Subscription upserted for user %s", req.UserID)
	return subscription, nil
}

func (m *SubscriptionManager) GetSubscription(ctx context.Context, userID string) (*Subscription, error) {
	var subscription Subscription
	err := m.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&subscription)
	if err != nil {
		return nil, err
	}

	// Check if expired and update status
	if subscription.Status == StatusActive && time.Now().After(subscription.ExpiresAt) {
		subscription.Status = StatusExpired
		m.collection.UpdateOne(ctx, bson.M{"user_id": userID}, bson.M{"$set": bson.M{"status": StatusExpired}})
	}

	// Get plan details
	if plan, err := m.planManager.GetByID(ctx, subscription.PlanID); err == nil {
		subscription.Plan = plan
	}

	return &subscription, nil
}

func (m *SubscriptionManager) CancelSubscription(ctx context.Context, userID string) error {
	subscription, err := m.GetSubscription(ctx, userID)
	if err != nil {
		return errors.New("subscription not found")
	}

	if subscription.Status != StatusActive {
		return errors.New("can only cancel active subscriptions")
	}

	_, err = m.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"status": StatusCancelled}},
	)
	return err
}
