package services

import (
	"context"
	"errors"
	"log"
	"subservice/core/models"
	"subservice/core/repository"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type SubscriptionService struct {
	subscriptionRepo *repository.SubscriptionRepository
	planRepo         *repository.PlanRepository
}

func NewSubscriptionService(subscriptionRepo *repository.SubscriptionRepository, planRepo *repository.PlanRepository) *SubscriptionService {
	return &SubscriptionService{
		subscriptionRepo: subscriptionRepo,
		planRepo:         planRepo,
	}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req *models.CreateSubscriptionRequest) (*models.Subscription, error) {
	// Get plan details first
	plan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	// Check if user already has a subscription
	existingSub, err := s.subscriptionRepo.GetByUserID(ctx, req.UserID)

	// Calculate new subscription dates
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

	subscription := &models.Subscription{
		UserID:    req.UserID,
		PlanID:    req.PlanID,
		Status:    models.StatusActive,
		StartDate: now,
		ExpiresAt: expiryDate,
	}

	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	if existingSub != nil {
		// User already has a subscription, update it (upsert behavior)
		log.Printf("Updating existing subscription for user %s", req.UserID)
		if err := s.subscriptionRepo.Upsert(ctx, subscription); err != nil {
			return nil, err
		}
	} else {
		// Create new subscription
		if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
			return nil, err
		}
	}

	// Populate plan details for response
	subscription.Plan = plan
	return subscription, nil
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, userID string) (*models.Subscription, error) {
	subscription, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if subscription has expired and update status if needed
	if subscription.Status == models.StatusActive && time.Now().After(subscription.ExpiresAt) {
		subscription.Status = models.StatusExpired

		// Update the status in database
		if err := s.subscriptionRepo.Update(ctx, userID, subscription); err != nil {
			log.Printf("Failed to update expired subscription status for user %s: %v", userID, err)
		}
	}

	// Get plan details
	plan, err := s.planRepo.GetByID(ctx, subscription.PlanID)
	if err == nil {
		subscription.Plan = plan
	}

	return subscription, nil
}

func (s *SubscriptionService) UpdateSubscription(ctx context.Context, userID string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	// Get existing subscription
	subscription, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("subscription not found")
	}

	// Check if subscription has expired and update status
	if subscription.Status == models.StatusActive && time.Now().After(subscription.ExpiresAt) {
		subscription.Status = models.StatusExpired
		s.subscriptionRepo.Update(ctx, userID, subscription)
		return nil, errors.New("cannot update expired subscription")
	}

	if subscription.Status != models.StatusActive {
		return nil, errors.New("can only update active subscriptions")
	}

	// Get new plan details
	newPlan, err := s.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	// Update subscription with new plan and reset dates
	now := time.Now()
	subscription.PlanID = req.PlanID
	subscription.Plan = newPlan
	subscription.StartDate = now // Reset start date for plan change

	// Recalculate expiry based on new plan
	switch newPlan.Duration {
	case "monthly":
		subscription.ExpiresAt = now.AddDate(0, 1, 0)
	case "yearly":
		subscription.ExpiresAt = now.AddDate(1, 0, 0)
	}

	if err := s.subscriptionRepo.Update(ctx, userID, subscription); err != nil {
		return nil, err
	}

	return subscription, nil
}

func (s *SubscriptionService) CancelSubscription(ctx context.Context, userID string) error {
	subscription, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return errors.New("subscription not found")
	}

	// Check if subscription has expired and update status
	if subscription.Status == models.StatusActive && time.Now().After(subscription.ExpiresAt) {
		subscription.Status = models.StatusExpired
		s.subscriptionRepo.Update(ctx, userID, subscription)
		return errors.New("subscription has already expired")
	}

	if subscription.Status != models.StatusActive {
		return errors.New("can only cancel active subscriptions")
	}

	subscription.Status = models.StatusCancelled
	return s.subscriptionRepo.Update(ctx, userID, subscription)
}

// New method for renewing cancelled subscriptions
func (s *SubscriptionService) RenewSubscription(ctx context.Context, userID string) (*models.Subscription, error) {
	subscription, err := s.subscriptionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("subscription not found")
	}

	if subscription.Status != models.StatusCancelled {
		return nil, errors.New("can only renew cancelled subscriptions")
	}

	// Get plan details
	plan, err := s.planRepo.GetByID(ctx, subscription.PlanID)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	// Renew with same plan but update created and expired times
	now := time.Now()
	subscription.Status = models.StatusActive
	subscription.StartDate = now
	subscription.CreatedAt = now // Update created time for renewal

	// Calculate new expiry date
	switch plan.Duration {
	case "monthly":
		subscription.ExpiresAt = now.AddDate(0, 1, 0)
	case "yearly":
		subscription.ExpiresAt = now.AddDate(1, 0, 0)
	}

	// Use upsert to update created time as well
	if err := s.subscriptionRepo.Upsert(ctx, subscription); err != nil {
		return nil, err
	}

	subscription.Plan = plan
	return subscription, nil
}
