package middleware

import (
	"github.com/gin-gonic/gin"
)

// SetReferrerPolicy sets the Referrer-Policy header
func SetReferrerPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}
