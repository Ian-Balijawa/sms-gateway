package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	ServerPort string
	ServerHost string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// SMS Provider configuration (egosms.co)
	SMSLiveURL     string
	SMSSandboxURL  string
	SMSUsername    string
	SMSPassword    string
	SMSSenderID    string
	SMSSandboxMode bool

	// API configuration
	JWTSecret string

	// Rate limiting
	RateLimitRPS int
}

var AppConfig *Config

func LoadConfig() error {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	AppConfig = &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "sms_gateway"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		SMSLiveURL:     getEnv("SMS_LIVE_URL", "https://www.egosms.co/api/v1/json/"),
		SMSSandboxURL:  getEnv("SMS_SANDBOX_URL", "http://sandbox.egosms.co/api/v1/json/"),
		SMSUsername:    getEnv("SMS_USERNAME", ""),
		SMSPassword:    getEnv("SMS_PASSWORD", ""),
		SMSSenderID:    getEnv("SMS_SENDER_ID", ""),
		SMSSandboxMode: getEnv("SMS_SANDBOX_MODE", "true") == "true",

		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),

		RateLimitRPS: getEnvAsInt("RATE_LIMIT_RPS", 100),
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

