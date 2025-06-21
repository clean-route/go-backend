package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/clean-route/go-backend/internal/logger"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		logger.Info("HTTP Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"response_size", blw.body.Len(),
			"request_id", c.GetString("request_id"),
		)

		// Log request body for debugging (only for non-GET requests)
		if c.Request.Method != "GET" && len(requestBody) > 0 {
			logger.Debug("Request Body",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"body", string(requestBody),
			)
		}

		// Log response body for errors
		if c.Writer.Status() >= 400 {
			logger.Error("HTTP Error Response",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", c.Writer.Status(),
				"response_body", blw.body.String(),
			)
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
