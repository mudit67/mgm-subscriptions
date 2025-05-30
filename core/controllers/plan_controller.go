package controllers

import (
	"net/http"
	"subservice/core/models"
	"subservice/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlanController struct {
	planManager *models.PlanManager
	validator   *validator.Validate
}

func NewPlanController(planManager *models.PlanManager) *PlanController {
	return &PlanController{
		planManager: planManager,
		validator:   validator.New(),
	}
}

func (c *PlanController) CreatePlan(ctx *gin.Context) {
	var plan models.Plan
	if err := ctx.ShouldBindJSON(&plan); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.validator.Struct(&plan); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.planManager.Create(ctx.Request.Context(), &plan); err != nil {
		utils.InternalErrorResponse(ctx, err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "Plan created successfully", plan)
}

func (c *PlanController) GetAllPlans(ctx *gin.Context) {
	plans, err := c.planManager.GetAll(ctx.Request.Context())
	if err != nil {
		utils.InternalErrorResponse(ctx, err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Plans retrieved successfully", plans)
}

func (c *PlanController) UpdatePlan(ctx *gin.Context) {
	id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid plan ID", err)
		return
	}

	var plan models.Plan
	if err := ctx.ShouldBindJSON(&plan); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.validator.Struct(&plan); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.planManager.Update(ctx.Request.Context(), id, &plan); err != nil {
		utils.InternalErrorResponse(ctx, err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Plan updated successfully", plan)
}

func (c *PlanController) DeletePlan(ctx *gin.Context) {
	id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid plan ID", err)
		return
	}

	if err := c.planManager.Delete(ctx.Request.Context(), id); err != nil {
		utils.InternalErrorResponse(ctx, err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Plan deleted successfully", nil)
}
