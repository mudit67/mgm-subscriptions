package repository

import (
	"context"
	"subservice/core/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SubscriptionRepository struct {
	collection *mongo.Collection
}

func NewSubscriptionRepository(db *mongo.Database) *SubscriptionRepository {
	repo := &SubscriptionRepository{
		collection: db.Collection("subscriptions"),
	}

	// Create unique index on user_id to ensure one subscription per user
	repo.createIndexes()
	return repo
}

func (r *SubscriptionRepository) createIndexes() {
	ctx := context.Background()

	// Create unique index on user_id
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: &options.IndexOptions{Unique: &[]bool{true}[0]},
	}

	r.collection.Indexes().CreateOne(ctx, indexModel)
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	subscription.ID = primitive.NewObjectID()
	subscription.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, subscription)
	return err
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, userID string, subscription *models.Subscription) error {
	// Don't update CreatedAt, only update the subscription data
	updateDoc := bson.M{
		"$set": bson.M{
			"plan_id":    subscription.PlanID,
			"status":     subscription.Status,
			"start_date": subscription.StartDate,
			"expires_at": subscription.ExpiresAt,
		},
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		updateDoc,
	)
	return err
}

func (r *SubscriptionRepository) Upsert(ctx context.Context, subscription *models.Subscription) error {
	subscription.CreatedAt = time.Now()

	filter := bson.M{"user_id": subscription.UserID}
	update := bson.M{
		"$set": bson.M{
			"plan_id":    subscription.PlanID,
			"status":     subscription.Status,
			"start_date": subscription.StartDate,
			"expires_at": subscription.ExpiresAt,
			"created_at": subscription.CreatedAt,
		},
		"$setOnInsert": bson.M{
			"_id": primitive.NewObjectID(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *SubscriptionRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

func (r *SubscriptionRepository) GetExpiredSubscriptions(ctx context.Context) ([]models.Subscription, error) {
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
		"status":     bson.M{"$ne": models.StatusExpired},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var subscriptions []models.Subscription
	if err = cursor.All(ctx, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.SubscriptionStatus) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status": status,
			},
		},
	)
	return err
}
