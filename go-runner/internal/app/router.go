package app

import (
	_ "go-runner/docs"
	"go-runner/internal/middleware"
	"go-runner/internal/project"
	"go-runner/internal/service"
	"go-runner/internal/system"
	"go-runner/internal/websocket"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(r *gin.Engine, db *gorm.DB) {
	// Global middleware
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.ErrorHandler())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.ErrorLogger())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.RateLimiter())
	r.Use(middleware.CORS())

	// Initialize service manager and websocket hub
	manager := service.NewManager(db)
	hub := websocket.NewHub()
	
	// Start websocket hub in goroutine
	go hub.Run()

	// Health check endpoint
	// @Summary      Health check
	// @Description  Check if the service is running
	// @Tags         health
	// @Accept       json
	// @Produce      json
	// @Success      200  {object}  map[string]interface{}  "Service status"
	// @Router       /health [get]
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "go-runner",
			"version": "1.0.0",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := r.Group("/api/v1")
	{
		// Project routes
		project.RegisterRoutes(api, db, manager, hub)
		
		// System monitoring routes
		system.RegisterRoutes(api, db)
	}

	// Root endpoint
	// @Summary      API Information
	// @Description  Get API information and available features
	// @Tags         info
	// @Accept       json
	// @Produce      json
	// @Success      200  {object}  map[string]interface{}  "API information"
	// @Router       / [get]
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Go Runner API - Microservice Management",
			"version": "1.0.0",
			"docs": "/swagger/index.html",
			"api": "/api/v1",
			"features": []string{
				"Microservice Management",
				"Real-time Logs",
				"Project Groups",
				"Health Monitoring",
				"WebSocket Support",
			},
		})
	})
}
