package controllers

import (
	"net/http"
	"subservice/core/models"
	"subservice/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UserController struct {
	userManager *models.UserManager
	validator   *validator.Validate
}

func NewUserController(userManager *models.UserManager) *UserController {
	return &UserController{
		userManager: userManager,
		validator:   validator.New(),
	}
}

func (c *UserController) Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	user, err := c.userManager.Register(ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Registration failed", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "User registered successfully", user)
}

func (c *UserController) Login(ctx *gin.Context) {
	var req models.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		utils.ValidationErrorResponse(ctx, err)
		return
	}

	loginResponse, err := c.userManager.Login(ctx.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "Login failed", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Login successful", loginResponse)
}
