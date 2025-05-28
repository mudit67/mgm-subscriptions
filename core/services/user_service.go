package services

import (
	"context"
	"errors"
	"subservice/core/models"
	"subservice/core/repository"
	"subservice/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	jwtExpiry string
}

func NewUserService(userRepo *repository.UserRepository, jwtSecret, jwtExpiry string) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if username already exists
	exists, err := s.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: req.Username,
		Name:     req.Name,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Parse JWT expiry
	expiry, err := time.ParseDuration(s.jwtExpiry)
	if err != nil {
		expiry = 24 * time.Hour // Default to 24 hours
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex(), s.jwtSecret, expiry)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
