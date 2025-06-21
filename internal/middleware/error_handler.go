package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/clean-route/go-backend/internal/errors"
	"github.com/clean-route/go-backend/internal/logger"
)

// ErrorHandlerMiddleware catches panics and converts them to proper HTTP responses
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error("Panic recovered",
				"error", err,
				"stack", string(debug.Stack()),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"request_id", c.GetString("request_id"),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal server error",
				"type":    "INTERNAL_ERROR",
			})
		} else {
			logger.Error("Panic recovered with non-string error",
				"error", recovered,
				"stack", string(debug.Stack()),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"request_id", c.GetString("request_id"),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal server error",
				"type":    "INTERNAL_ERROR",
			})
		}
	})
}

// ErrorResponseMiddleware handles custom application errors
func ErrorResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Check if it's our custom AppError
			if appErr := errors.GetAppError(err); appErr != nil {
				logger.Error("Application error",
					"type", string(appErr.Type),
					"message", appErr.Message,
					"status_code", appErr.StatusCode,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"request_id", c.GetString("request_id"),
					"context", appErr.Context,
				)

				c.JSON(appErr.StatusCode, gin.H{
					"success": false,
					"error":   appErr.Message,
					"type":    appErr.Type,
					"context": appErr.Context,
				})
				return
			}

			// Handle generic errors
			logger.Error("Generic error",
				"error", err.Error(),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"request_id", c.GetString("request_id"),
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Internal server error",
				"type":    "INTERNAL_ERROR",
			})
		}
	}
}
