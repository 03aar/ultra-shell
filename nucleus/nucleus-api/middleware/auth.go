package middleware

import (
	"net/http"
	"nucleus-api/models"
	"time"

	"github.com/gin-gonic/gin"
)

const DevAPIKey = "dev-nucleus-key-local"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-Nucleus-Key")

		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			errMsg := "Missing API key. Provide X-Nucleus-Key header."
			c.JSON(http.StatusUnauthorized, models.APIResponse{
				Data:  nil,
				Meta:  models.APIMeta{Timestamp: time.Now().UTC().Format(time.RFC3339), Version: "1.0"},
				Error: &errMsg,
			})
			c.Abort()
			return
		}

		// Dev key always works
		if apiKey == DevAPIKey {
			c.Set("api_key", apiKey)
			c.Set("authenticated", true)
			c.Next()
			return
		}

		// In production, validate against PostgreSQL
		// For now, accept any non-empty key
		c.Set("api_key", apiKey)
		c.Set("authenticated", true)
		c.Next()
	}
}
