package controllers

import (
	"net/http"
	"subservice/core/models"
	"subservice/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SubscriptionController struct {
	subscriptionManager *models.SubscriptionManager
	validator           *validator.Validate
}

func NewSubscriptionController(subscriptionManager *models.SubscriptionManager) *SubscriptionController {
	return &SubscriptionController{
		subscriptionManager: subscriptionManager,
		validator:           validator.New(),
	}
}

func (c *SubscriptionController) UpsertSubscription(ctx *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	subscription, err := c.subscriptionManager.UpsertSubscription(ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Failed to process subscription", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Subscription processed successfully", subscription)
}

func (c *SubscriptionController) GetSubscription(ctx *gin.Context) {
	userID := ctx.Param("userId")
	if userID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	subscription, err := c.subscriptionManager.GetSubscription(ctx.Request.Context(), userID)
	if err != nil {
		utils.NotFoundResponse(ctx, "Subscription not found")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Subscription retrieved successfully", subscription)
}

func (c *SubscriptionController) CancelSubscription(ctx *gin.Context) {
	userID := ctx.Param("userId")
	if userID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	if err := c.subscriptionManager.CancelSubscription(ctx.Request.Context(), userID); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Failed to cancel subscription", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Subscription cancelled successfully", nil)
}
