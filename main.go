// There is no car mode input -- it is by default driving-traffic

package main

import (
	"log"
	"net/http"

	"github.com/clean-route/go-backend/internal/config"
	"github.com/clean-route/go-backend/internal/handlers"
	"github.com/clean-route/go-backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	port        = "9000"
	serviceName = "clean-route-service"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.Default()

	// Add middleware
	router.Use(cors.Default())
	router.Use(middleware.SetReferrerPolicy())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": serviceName,
		})
	})

	// Backward compatibility endpoints (original endpoints)
	router.POST("/route", handlers.FindRoute)
	router.POST("/all-routes", handlers.FindAllRoutes)

	// API routes (new versioned endpoints)
	api := router.Group("/api/v1")
	{
		// Route planning endpoints
		api.POST("/route", handlers.FindRoute)
		api.POST("/routes", handlers.FindAllRoutes)

		// Weather and air quality endpoints
		api.GET("/weather", handlers.GetWeatherData)
		api.GET("/aqi", handlers.GetAQIData)
		api.POST("/predict/pm25", handlers.GetPredictedPM25)
	}

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
