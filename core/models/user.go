package models

import (
	"context"
	"errors"
	"time"

	"subservice/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username string             `json:"username" bson:"username" validate:"required,min=3"`
	Name     string             `json:"name" bson:"name" validate:"required"`
	Password string             `json:"-" bson:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UserManager struct {
	collection *mongo.Collection
	jwtSecret  string
	jwtExpiry  string
}

func NewUserManager(db *mongo.Database, jwtSecret, jwtExpiry string) *UserManager {
	manager := &UserManager{
		collection: db.Collection("users"),
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
	}
	manager.createIndexes()
	return manager
}

func (m *UserManager) createIndexes() {
	ctx := context.Background()
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: &options.IndexOptions{Unique: &[]bool{true}[0]},
	}
	m.collection.Indexes().CreateOne(ctx, indexModel)
}

func (m *UserManager) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	// Check if username exists
	count, err := m.collection.CountDocuments(ctx, bson.M{"username": req.Username})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:       primitive.NewObjectID(),
		Username: req.Username,
		Name:     req.Name,
		Password: string(hashedPassword),
	}

	_, err = m.collection.InsertOne(ctx, user)
	return user, err
}

func (m *UserManager) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	var user User
	err := m.collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT token
	expiry, _ := time.ParseDuration(m.jwtExpiry)
	if expiry == 0 {
		expiry = 24 * time.Hour
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Username, m.jwtSecret, expiry)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: token, User: user}, nil
}

func (m *UserManager) GetByID(ctx context.Context, userID string) (*User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user User
	err = m.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	return &user, err
}
