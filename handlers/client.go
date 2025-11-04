package handlers

import (
	"net/http"
	"sms-gateway/database"
	"sms-gateway/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ClientHandler struct{}

func NewClientHandler() *ClientHandler {
	return &ClientHandler{}
}

// CreateClient creates a new API client (admin only)
func (h *ClientHandler) CreateClient(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		RateLimit    int    `json:"rate_limit"`
		DailyLimit   int    `json:"daily_limit"`
		MonthlyLimit int    `json:"monthly_limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SMSResponse{
			Success: false,
			Message: "Invalid request payload",
			Error:   err.Error(),
		})
		return
	}

	// Check if email already exists
	var existingClient models.APIClient
	if err := database.DB.Where("email = ?", req.Email).First(&existingClient).Error; err == nil {
		c.JSON(http.StatusConflict, models.SMSResponse{
			Success: false,
			Message: "Client with this email already exists",
		})
		return
	}

	// Generate API key and secret
	apiKey := uuid.New().String()
	apiSecret := uuid.New().String()

	// Hash the API secret
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(apiSecret), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to generate credentials",
			Error:   err.Error(),
		})
		return
	}

	// Set defaults
	rateLimit := req.RateLimit
	if rateLimit == 0 {
		rateLimit = 100
	}
	dailyLimit := req.DailyLimit
	if dailyLimit == 0 {
		dailyLimit = 10000
	}
	monthlyLimit := req.MonthlyLimit
	if monthlyLimit == 0 {
		monthlyLimit = 300000
	}

	// Create client
	client := models.APIClient{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		APIKey:       apiKey,
		APISecret:    string(hashedSecret),
		IsActive:     true,
		RateLimit:    rateLimit,
		DailyLimit:   dailyLimit,
		MonthlyLimit: monthlyLimit,
		LastReset:    time.Now(),
	}

	if err := database.DB.Create(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to create client",
			Error:   err.Error(),
		})
		return
	}

	// Return response with credentials (only shown once)
	c.JSON(http.StatusCreated, models.SMSResponse{
		Success: true,
		Message: "Client created successfully",
		Data: map[string]interface{}{
			"client_id":   client.ID,
			"name":        client.Name,
			"email":       client.Email,
			"api_key":     apiKey,
			"api_secret":  apiSecret, // Only shown on creation
			"rate_limit":  client.RateLimit,
			"daily_limit": client.DailyLimit,
			"monthly_limit": client.MonthlyLimit,
			"warning":     "Save these credentials securely. The API secret will not be shown again.",
		},
	})
}

// ListClients lists all API clients (admin only)
func (h *ClientHandler) ListClients(c *gin.Context) {
	var clients []models.APIClient
	
	query := database.DB

	// Filter by active status
	if isActive := c.Query("is_active"); isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}

	if err := query.Find(&clients).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to retrieve clients",
			Error:   err.Error(),
		})
		return
	}

	// Remove sensitive information
	for i := range clients {
		clients[i].APISecret = ""
		clients[i].APIKeyHash = ""
	}

	c.JSON(http.StatusOK, models.SMSResponse{
		Success: true,
		Message: "Clients retrieved successfully",
		Data:    clients,
	})
}

// UpdateClient updates a client's information (admin only)
func (h *ClientHandler) UpdateClient(c *gin.Context) {
	clientID := c.Param("id")
	var req struct {
		Name         *string `json:"name"`
		IsActive     *bool   `json:"is_active"`
		RateLimit    *int    `json:"rate_limit"`
		DailyLimit   *int    `json:"daily_limit"`
		MonthlyLimit *int    `json:"monthly_limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SMSResponse{
			Success: false,
			Message: "Invalid request payload",
			Error:   err.Error(),
		})
		return
	}

	var client models.APIClient
	if err := database.DB.Where("id = ?", clientID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, models.SMSResponse{
			Success: false,
			Message: "Client not found",
		})
		return
	}

	// Update fields
	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}
	if req.RateLimit != nil {
		client.RateLimit = *req.RateLimit
	}
	if req.DailyLimit != nil {
		client.DailyLimit = *req.DailyLimit
	}
	if req.MonthlyLimit != nil {
		client.MonthlyLimit = *req.MonthlyLimit
	}

	if err := database.DB.Save(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to update client",
			Error:   err.Error(),
		})
		return
	}

	// Remove sensitive information
	client.APISecret = ""
	client.APIKeyHash = ""

	c.JSON(http.StatusOK, models.SMSResponse{
		Success: true,
		Message: "Client updated successfully",
		Data:    client,
	})
}

// ResetClientUsage resets a client's usage counters (admin only)
func (h *ClientHandler) ResetClientUsage(c *gin.Context) {
	clientID := c.Param("id")

	var client models.APIClient
	if err := database.DB.Where("id = ?", clientID).First(&client).Error; err != nil {
		c.JSON(http.StatusNotFound, models.SMSResponse{
			Success: false,
			Message: "Client not found",
		})
		return
	}

	client.DailyUsage = 0
	client.MonthlyUsage = 0
	client.LastReset = time.Now()

	if err := database.DB.Save(&client).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to reset usage",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SMSResponse{
		Success: true,
		Message: "Client usage reset successfully",
	})
}

