package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth is a middleware that validates X-API-Key header
// Returns 401 if header is missing, 403 if key is wrong
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedKey := os.Getenv("API_KEY")

		// If API_KEY is not set, skip authentication (dev mode)
		if expectedKey == "" {
			c.Next()
			return
		}

		providedKey := c.GetHeader("X-API-Key")

		if providedKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "X-API-Key header is required",
			})
			c.Abort()
			return
		}

		if providedKey != expectedKey {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
