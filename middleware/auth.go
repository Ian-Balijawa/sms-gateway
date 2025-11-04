package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"github.com/Ian-Balijawa/sms-gateway/database"
	"github.com/Ian-Balijawa/sms-gateway/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// APIKeyAuth middleware validates API key and secret from request headers
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract API key from header
		apiKey := c.GetHeader("X-API-Key")
		apiSecret := c.GetHeader("X-API-Secret")

		if apiKey == "" || apiSecret == "" {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Missing API credentials",
				Error:   "X-API-Key and X-API-Secret headers are required",
			})
			c.Abort()
			return
		}

		// Find client by API key
		var client models.APIClient
		if err := database.DB.Where("api_key = ?", apiKey).First(&client).Error; err != nil {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Invalid API credentials",
				Error:   "API key not found",
			})
			c.Abort()
			return
		}

		// Verify API secret
		if err := bcrypt.CompareHashAndPassword([]byte(client.APISecret), []byte(apiSecret)); err != nil {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Invalid API credentials",
				Error:   "API secret mismatch",
			})
			c.Abort()
			return
		}

		// Check if client is active
		if !client.IsActive {
			c.JSON(http.StatusForbidden, models.SMSResponse{
				Success: false,
				Message: "API client is inactive",
				Error:   "Your API access has been suspended",
			})
			c.Abort()
			return
		}

		// Check rate limits (simplified - in production, use Redis for distributed rate limiting)
		if client.DailyUsage >= client.DailyLimit {
			c.JSON(http.StatusTooManyRequests, models.SMSResponse{
				Success: false,
				Message: "Daily limit exceeded",
				Error:   "You have reached your daily SMS limit",
			})
			c.Abort()
			return
		}

		if client.MonthlyUsage >= client.MonthlyLimit {
			c.JSON(http.StatusTooManyRequests, models.SMSResponse{
				Success: false,
				Message: "Monthly limit exceeded",
				Error:   "You have reached your monthly SMS limit",
			})
			c.Abort()
			return
		}

		// Store client in context for use in handlers
		c.Set("client", client)
		c.Set("client_id", client.ID)

		c.Next()
	}
}

// BasicAuth middleware for admin endpoints (optional)
func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Authorization required",
			})
			c.Abort()
			return
		}

		// Extract and validate Basic Auth
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Basic" {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Invalid authorization header",
			})
			c.Abort()
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Invalid authorization header",
			})
			c.Abort()
			return
		}

		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Invalid credentials format",
			})
			c.Abort()
			return
		}

		// In production, validate against database or config
		// For now, using constant time comparison
		expectedUser := getEnv("ADMIN_USER", "admin")
		expectedPass := getEnv("ADMIN_PASSWORD", "admin")

		if subtle.ConstantTimeCompare([]byte(credentials[0]), []byte(expectedUser)) != 1 ||
			subtle.ConstantTimeCompare([]byte(credentials[1]), []byte(expectedPass)) != 1 {
			c.JSON(http.StatusUnauthorized, models.SMSResponse{
				Success: false,
				Message: "Invalid credentials",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

