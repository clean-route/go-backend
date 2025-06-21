package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/clean-route/go-backend/internal/errors"
	"github.com/clean-route/go-backend/internal/logger"
	"github.com/clean-route/go-backend/internal/models"
	"github.com/clean-route/go-backend/internal/services"
)

var routeService = services.NewRouteService()

// FindRoute handles single route requests
func FindRoute(c *gin.Context) {
	var req models.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request format for FindRoute",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
		)

		appErr := errors.NewValidationError("Invalid request format", err)
		c.Error(appErr)
		return
	}

	logger.Info("Processing single route request",
		"request_id", c.GetString("request_id"),
		"source", req.Source,
		"destination", req.Destination,
		"mode", req.Mode,
		"route_preference", req.RoutePreference,
	)

	result, err := routeService.FindSingleRoute(req)
	if err != nil {
		logger.Error("Failed to find single route",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"request", req,
		)

		appErr := errors.NewInternalError("Failed to find route", err)
		c.Error(appErr)
		return
	}

	logger.Info("Successfully found single route",
		"request_id", c.GetString("request_id"),
	)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// FindAllRoutes handles requests for all route types
func FindAllRoutes(c *gin.Context) {
	var req models.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request format for FindAllRoutes",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
		)

		appErr := errors.NewValidationError("Invalid request format", err)
		c.Error(appErr)
		return
	}

	logger.Info("Processing all routes request",
		"request_id", c.GetString("request_id"),
		"source", req.Source,
		"destination", req.Destination,
		"mode", req.Mode,
		"route_preference", req.RoutePreference,
	)

	result, err := routeService.FindAllRoutes(req)
	if err != nil {
		logger.Error("Failed to find all routes",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"request", req,
		)

		appErr := errors.NewInternalError("Failed to find routes", err)
		c.Error(appErr)
		return
	}

	logger.Info("Successfully found all routes",
		"request_id", c.GetString("request_id"),
	)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetWeatherData handles weather data requests
func GetWeatherData(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		logger.Warn("Missing required query parameters for weather data",
			"request_id", c.GetString("request_id"),
			"lat", latStr,
			"lon", lonStr,
		)

		appErr := errors.NewValidationError("Missing required query parameters: lat and lon", nil)
		c.Error(appErr)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		logger.Error("Invalid latitude parameter",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"lat", latStr,
		)

		appErr := errors.NewValidationError("Invalid latitude parameter", err)
		c.Error(appErr)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		logger.Error("Invalid longitude parameter",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"lon", lonStr,
		)

		appErr := errors.NewValidationError("Invalid longitude parameter", err)
		c.Error(appErr)
		return
	}

	logger.Info("Fetching weather data",
		"request_id", c.GetString("request_id"),
		"lat", lat,
		"lon", lon,
	)

	location := []float64{lon, lat}
	weatherData := services.FetchWeatherData(location)

	logger.Info("Successfully fetched weather data",
		"request_id", c.GetString("request_id"),
	)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    weatherData,
	})
}

// GetAQIData handles air quality data requests
func GetAQIData(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")

	if latStr == "" || lonStr == "" {
		logger.Warn("Missing required query parameters for AQI data",
			"request_id", c.GetString("request_id"),
			"lat", latStr,
			"lon", lonStr,
		)

		appErr := errors.NewValidationError("Missing required query parameters: lat and lon", nil)
		c.Error(appErr)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		logger.Error("Invalid latitude parameter for AQI",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"lat", latStr,
		)

		appErr := errors.NewValidationError("Invalid latitude parameter", err)
		c.Error(appErr)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		logger.Error("Invalid longitude parameter for AQI",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"lon", lonStr,
		)

		appErr := errors.NewValidationError("Invalid longitude parameter", err)
		c.Error(appErr)
		return
	}

	logger.Info("Fetching AQI data",
		"request_id", c.GetString("request_id"),
		"lat", lat,
		"lon", lon,
	)

	location := []float64{lon, lat}
	aqiValue, err := services.FetchAQIData(location, 0) // Default delay code
	if err != nil {
		logger.Error("Failed to fetch AQI data",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"lat", lat,
			"lon", lon,
		)

		appErr := errors.NewExternalError("Failed to fetch AQI data", err)
		c.Error(appErr)
		return
	}

	logger.Info("Successfully fetched AQI data",
		"request_id", c.GetString("request_id"),
		"aqi_value", aqiValue,
	)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"aqi": aqiValue,
		},
	})
}

// GetPredictedPM25 handles PM2.5 prediction requests
func GetPredictedPM25(c *gin.Context) {
	var req models.PM25PredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request format for PM2.5 prediction",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
		)

		appErr := errors.NewValidationError("Invalid request format", err)
		c.Error(appErr)
		return
	}

	logger.Info("Processing PM2.5 prediction request",
		"request_id", c.GetString("request_id"),
		"features_count", len(req.Features),
	)

	predictions, err := services.GetPredictedPm25(req.Features)
	if err != nil {
		logger.Error("Failed to get PM2.5 predictions",
			"error", err.Error(),
			"request_id", c.GetString("request_id"),
			"request", req,
		)

		appErr := errors.NewInternalError("Failed to get PM2.5 predictions", err)
		c.Error(appErr)
		return
	}

	logger.Info("Successfully generated PM2.5 predictions",
		"request_id", c.GetString("request_id"),
		"predictions_count", len(predictions),
	)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: gin.H{
			"predictions": predictions,
		},
	})
}
