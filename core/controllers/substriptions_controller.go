package controllers

import (
	"net/http"
	"subservice/core/models"
	"subservice/core/services"
	"subservice/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SubscriptionController struct {
	subscriptionService *services.SubscriptionService
	validator           *validator.Validate
}

func NewSubscriptionController(subscriptionService *services.SubscriptionService) *SubscriptionController {
	return &SubscriptionController{
		subscriptionService: subscriptionService,
		validator:           validator.New(),
	}
}

func (c *SubscriptionController) CreateSubscription(ctx *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	subscription, err := c.subscriptionService.CreateSubscription(ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Failed to create subscription", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "Subscription created successfully", subscription)
}

func (c *SubscriptionController) GetSubscription(ctx *gin.Context) {
	userID := ctx.Param("userId")
	if userID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	subscription, err := c.subscriptionService.GetSubscription(ctx.Request.Context(), userID)
	if err != nil {
		utils.NotFoundResponse(ctx, "Subscription not found")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Subscription retrieved successfully", subscription)
}

func (c *SubscriptionController) RenewSubscription(ctx *gin.Context) {
	userID := ctx.Param("userId")
	if userID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	subscription, err := c.subscriptionService.RenewSubscription(ctx.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Failed to renew subscription", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Subscription renewed successfully", subscription)
}

func (c *SubscriptionController) UpdateSubscription(ctx *gin.Context) {
	userID := ctx.Param("userId")
	if userID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	subscription, err := c.subscriptionService.UpdateSubscription(ctx.Request.Context(), userID, &req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update subscription", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Subscription updated successfully", subscription)
}

func (c *SubscriptionController) CancelSubscription(ctx *gin.Context) {
	userID := ctx.Param("userId")
	if userID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	if err := c.subscriptionService.CancelSubscription(ctx.Request.Context(), userID); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Failed to cancel subscription", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Subscription cancelled successfully", nil)
}
