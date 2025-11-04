package utils

import (
	"log"
		"github.com/Ian-Balijawa/sms-gateway/database"
	"time"
	"github.com/Ian-Balijawa/sms-gateway/models"
)

// ResetDailyUsage resets daily usage for all clients
// Should be called daily via a cron job or scheduler
func ResetDailyUsage() error {
	result := database.DB.Model(&models.APIClient{}).
		Where("last_reset < ? OR last_reset IS NULL", time.Now().AddDate(0, 0, -1)).
		Updates(map[string]interface{}{
			"daily_usage": 0,
			"last_reset":  time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	log.Printf("Reset daily usage for %d clients", result.RowsAffected)
	return nil
}

// ResetMonthlyUsage resets monthly usage for all clients
// Should be called monthly via a cron job or scheduler
func ResetMonthlyUsage() error {
	result := database.DB.Model(&models.APIClient{}).
		Where("last_reset < ? OR last_reset IS NULL", time.Now().AddDate(0, -1, 0)).
		Updates(map[string]interface{}{
			"monthly_usage": 0,
			"last_reset":   time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	log.Printf("Reset monthly usage for %d clients", result.RowsAffected)
	return nil
}

// StartUsageResetScheduler starts background goroutines to reset usage periodically
func StartUsageResetScheduler() {
	// Reset daily usage at midnight
	go func() {
		for {
			now := time.Now()
			nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			duration := nextMidnight.Sub(now)

			time.Sleep(duration)
			if err := ResetDailyUsage(); err != nil {
				log.Printf("Error resetting daily usage: %v", err)
			}
		}
	}()

	// Reset monthly usage on the first day of each month
	go func() {
		for {
			now := time.Now()
			nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
			duration := nextMonth.Sub(now)

			time.Sleep(duration)
			if err := ResetMonthlyUsage(); err != nil {
				log.Printf("Error resetting monthly usage: %v", err)
			}
		}
	}()

	log.Println("Usage reset scheduler started")
}

