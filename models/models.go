package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIClient represents a client application that can use the SMS gateway
type APIClient struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Client identification
	Name        string `gorm:"not null" json:"name"`
	Email       string `gorm:"uniqueIndex;not null" json:"email"`
	APIKey      string `gorm:"uniqueIndex;not null" json:"api_key"`
	APISecret   string `gorm:"not null" json:"-"` // Never exposed in JSON
	APIKeyHash  string `gorm:"not null" json:"-"` // Hash of API key for validation

	// Status and limits
	IsActive    bool `gorm:"default:true" json:"is_active"`
	RateLimit   int  `gorm:"default:100" json:"rate_limit"` // Requests per second
	DailyLimit  int  `gorm:"default:10000" json:"daily_limit"`
	MonthlyLimit int `gorm:"default:300000" json:"monthly_limit"`

	// Usage tracking
	DailyUsage   int       `gorm:"default:0" json:"daily_usage"`
	MonthlyUsage int       `gorm:"default:0" json:"monthly_usage"`
	LastReset    time.Time `json:"last_reset"`
}

// BeforeCreate hook to generate UUID before creating
func (client *APIClient) BeforeCreate(tx *gorm.DB) error {
	if client.ID == uuid.Nil {
		client.ID = uuid.New()
	}
	if client.LastReset.IsZero() {
		client.LastReset = time.Now()
	}
	return nil
}

// SMSLog represents a log entry for each SMS sent
type SMSLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	// Client information
	ClientID uuid.UUID `gorm:"type:uuid;not null;index" json:"client_id"`
	Client   APIClient `gorm:"foreignKey:ClientID" json:"client,omitempty"`

	// SMS details
	Recipient  string `gorm:"not null" json:"recipient"`
	Message    string `gorm:"not null" json:"message"`
	SenderID   string `json:"sender_id"`
	Priority   string `gorm:"default:1" json:"priority"`

	// Status
	Status     string `gorm:"not null" json:"status"` // "pending", "sent", "failed"
	ProviderStatus string `json:"provider_status"`    // Status from SMS provider
	ProviderMessage string `json:"provider_message"`  // Message from SMS provider
	Error      string `json:"error,omitempty"`

	// Metadata
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
}

// BeforeCreate hook to generate UUID before creating
func (log *SMSLog) BeforeCreate(tx *gorm.DB) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	return nil
}

// SMSRequest represents the incoming SMS request payload
type SMSRequest struct {
	Number   string `json:"number" binding:"required"`
	Message  string `json:"message" binding:"required"`
	SenderID string `json:"senderid,omitempty"`
	Priority string `json:"priority,omitempty"`
}

// BulkSMSRequest represents multiple SMS requests
type BulkSMSRequest struct {
	Messages []SMSRequest `json:"messages" binding:"required,min=1,dive"`
}

// SMSResponse represents the API response
type SMSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ClientStats represents usage statistics for a client
type ClientStats struct {
	ClientID      uuid.UUID `json:"client_id"`
	DailyUsage    int       `json:"daily_usage"`
	MonthlyUsage  int       `json:"monthly_usage"`
	DailyLimit    int       `json:"daily_limit"`
	MonthlyLimit  int       `json:"monthly_limit"`
	IsActive      bool      `json:"is_active"`
}

