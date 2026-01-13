package main

import (
	"log"
	"os"

	"checkopen/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// Create handler
	statusHandler := handler.NewStatusHandler()

	// Routes
	r.GET("/health", statusHandler.HealthCheck)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/status/:restaurantId", statusHandler.GetStatus)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting server on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
