package services

import (
	"context"
	"subservice/core/models"
	"subservice/core/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlanService struct {
	planRepo *repository.PlanRepository
}

func NewPlanService(planRepo *repository.PlanRepository) *PlanService {
	return &PlanService{
		planRepo: planRepo,
	}
}

func (s *PlanService) CreatePlan(ctx context.Context, plan *models.Plan) error {
	return s.planRepo.Create(ctx, plan)
}

func (s *PlanService) GetAllPlans(ctx context.Context) ([]models.Plan, error) {
	return s.planRepo.GetAll(ctx)
}

func (s *PlanService) GetPlanByID(ctx context.Context, id primitive.ObjectID) (*models.Plan, error) {
	return s.planRepo.GetByID(ctx, id)
}

func (s *PlanService) UpdatePlan(ctx context.Context, id primitive.ObjectID, plan *models.Plan) error {
	return s.planRepo.Update(ctx, id, plan)
}

func (s *PlanService) DeletePlan(ctx context.Context, id primitive.ObjectID) error {
	return s.planRepo.Delete(ctx, id)
}
