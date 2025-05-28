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
	"subservice/core/models"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	// Connect to MongoDB with retry logic
	mongoDB, err := database.MongoDbInit(cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB after retries:", err)
	}
	defer mongoDB.Close()

	// Start background reconnection monitoring
	go mongoDB.BackgroundReconnect(cfg.MongoURI, cfg.DatabaseName)

	// Initialize managers
	userManager := models.NewUserManager(mongoDB.Database, cfg.JWTSecret, cfg.JWTExpiry)
	planManager := models.NewPlanManager(mongoDB.Database)
	subscriptionManager := models.NewSubscriptionManager(mongoDB.Database, planManager)

	// Initialize controllers
	userController := controllers.NewUserController(userManager)
	planController := controllers.NewPlanController(planManager)
	subscriptionController := controllers.NewSubscriptionController(subscriptionManager)

	// Setup router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(gin.Logger(), gin.Recovery())

	// CORS middleware
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

	// Static files
	router.Static("/css", "./frontend/css")
	router.Static("/js", "./frontend/js")
	router.GET("/", func(c *gin.Context) { c.File("./frontend/index.html") })
	router.GET("/dashboard", func(c *gin.Context) { c.File("./frontend/dashboard.html") })
	router.GET("/admin", func(c *gin.Context) { c.File("./frontend/admin.html") })

	// Health check with database ping
	router.GET("/health", func(c *gin.Context) {
		if err := mongoDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
		})
	})

	// API routes
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", userController.Register)
			auth.POST("/login", userController.Login)
		}

		api.GET("/plans", planController.GetAllPlans)

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.POST("/subscriptions", subscriptionController.UpsertSubscription)
			protected.PUT("/subscriptions", subscriptionController.UpsertSubscription)
			protected.GET("/subscriptions/:userId", subscriptionController.GetSubscription)
			protected.DELETE("/subscriptions/:userId", subscriptionController.CancelSubscription)

			adminOnly := protected.Group("/")
			adminOnly.Use(middleware.AdminMiddleware())
			{
				adminOnly.POST("/plans", planController.CreatePlan)
				adminOnly.PUT("/plans/:id", planController.UpdatePlan)
				adminOnly.DELETE("/plans/:id", planController.DeletePlan)
			}
		}
	}

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}
