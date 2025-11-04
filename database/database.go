package database

import (
	"fmt"
	"log"
	"os"
	"sms-gateway/config"
	"sms-gateway/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	var err error
	var dialector gorm.Dialector

	// Use SQLite for development, PostgreSQL for production
	// You can switch based on environment variable
	dbType := getEnv("DB_TYPE", "sqlite")

	if dbType == "postgres" {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			config.AppConfig.DBHost,
			config.AppConfig.DBUser,
			config.AppConfig.DBPassword,
			config.AppConfig.DBName,
			config.AppConfig.DBPort,
			config.AppConfig.DBSSLMode,
		)
		dialector = postgres.Open(dsn)
	} else {
		// SQLite for development
		dialector = sqlite.Open("sms_gateway.db")
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	err = DB.AutoMigrate(
		&models.APIClient{},
		&models.SMSLog{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

