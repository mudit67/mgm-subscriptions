package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subservice/core/config"
	"subservice/core/controllers"
	"subservice/core/database"
	"subservice/core/middleware"
	"subservice/core/repository"
	"subservice/core/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to MongoDB
	mongoDB, err := database.MongoDbInit(cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoDB.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(mongoDB.Database)
	planRepo := repository.NewPlanRepository(mongoDB.Database)
	subscriptionRepo := repository.NewSubscriptionRepository(mongoDB.Database)

	// Initialize services
	userService := services.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	planService := services.NewPlanService(planRepo)
	subscriptionService := services.NewSubscriptionService(subscriptionRepo, planRepo)

	// Initialize controllers
	userController := controllers.NewUserController(userService)
	planController := controllers.NewPlanController(planService)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
		})
	})

	// CORS middleware for frontend
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files from frontend directory
	router.Static("/css", "./frontend/css")
	router.Static("/js", "./frontend/js")
	router.Static("/assets", "./frontend/assets")

	// Serve index.html on root path
	router.GET("/", func(c *gin.Context) {
		c.File("./frontend/index.html")
	})

	// Serve dashboard.html
	router.GET("/dashboard", func(c *gin.Context) {
		c.File("./frontend/dashboard.html")
	})

	// Change your routes to use /api prefix
	api := router.Group("/api")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", userController.Register)
			auth.POST("/login", userController.Login)
		}

		// Public plans endpoint
		api.GET("/plans", planController.GetAllPlans)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Subscription routes
			protected.POST("/subscriptions", subscriptionController.CreateSubscription)
			protected.GET("/subscriptions/:userId", subscriptionController.GetSubscription)
			protected.PUT("/subscriptions/:userId", subscriptionController.UpdateSubscription)
			protected.DELETE("/subscriptions/:userId", subscriptionController.CancelSubscription)
			protected.POST("/subscriptions/:userId/renew", subscriptionController.RenewSubscription)

			// Admin plan routes
			protected.POST("/plans", planController.CreatePlan)
			protected.PUT("/plans/:id", planController.UpdatePlan)
			protected.DELETE("/plans/:id", planController.DeletePlan)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
