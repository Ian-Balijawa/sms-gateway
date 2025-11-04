package handlers

import (
	"net/http"
	"sms-gateway/database"
	"sms-gateway/models"
	"sms-gateway/service"
	"sms-gateway/utils"

	"github.com/gin-gonic/gin"
)

type SMSHandler struct {
	smsProvider *service.SMSProvider
}

func NewSMSHandler() *SMSHandler {
	return &SMSHandler{
		smsProvider: service.NewSMSProvider(),
	}
}

// SendSingleSMS handles sending a single SMS
func (h *SMSHandler) SendSingleSMS(c *gin.Context) {
	var req models.SMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SMSResponse{
			Success: false,
			Message: "Invalid request payload",
			Error:   err.Error(),
		})
		return
	}

	// Validate phone number
	if !utils.ValidatePhone(req.Number) {
		c.JSON(http.StatusBadRequest, models.SMSResponse{
			Success: false,
			Message: "Invalid phone number",
			Error:   "Phone number format is invalid",
		})
		return
	}

	// Get client from context (set by auth middleware)
	client, exists := c.Get("client")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Client not found in context",
		})
		return
	}
	apiClient := client.(models.APIClient)
	clientID := apiClient.ID

	// Send SMS via provider
	responses, err := h.smsProvider.SendSMS([]models.SMSRequest{req}, apiClient.Name)
	if err != nil {
		// Log error
		smsLog := models.SMSLog{
			ClientID:       clientID,
			Recipient:      req.Number,
			Message:        req.Message,
			SenderID:       req.SenderID,
			Priority:       req.Priority,
			Status:         "failed",
			ProviderStatus: "error",
			Error:          err.Error(),
			IPAddress:      c.ClientIP(),
			UserAgent:      c.GetHeader("User-Agent"),
		}
		database.DB.Create(&smsLog)

		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to send SMS",
			Error:   err.Error(),
		})
		return
	}

	// Determine status from provider response
	status := "sent"
	providerStatus := "Success"
	providerMessage := "SMS sent successfully"
	errorMsg := ""

	if len(responses) > 0 {
		resp := responses[0]
		if resp.Status != "Success" && resp.Status != "success" {
			status = "failed"
			providerStatus = resp.Status
			providerMessage = resp.Message
			errorMsg = resp.Message
		} else {
			providerMessage = resp.Message
		}
	}

	// Log SMS
	smsLog := models.SMSLog{
		ClientID:       clientID,
		Recipient:      utils.FormatPhone(req.Number),
		Message:        req.Message,
		SenderID:       req.SenderID,
		Priority:       req.Priority,
		Status:         status,
		ProviderStatus: providerStatus,
		ProviderMessage: providerMessage,
		Error:          errorMsg,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.GetHeader("User-Agent"),
	}
	database.DB.Create(&smsLog)

	// Update client usage
	if status == "sent" {
		apiClient.DailyUsage++
		apiClient.MonthlyUsage++
		database.DB.Save(&apiClient)
	}

	// Return response
	if status == "sent" {
		c.JSON(http.StatusOK, models.SMSResponse{
			Success: true,
			Message: "SMS sent successfully",
			Data: map[string]interface{}{
				"log_id":    smsLog.ID,
				"recipient": smsLog.Recipient,
				"status":    status,
				"provider_response": map[string]string{
					"status":  providerStatus,
					"message": providerMessage,
				},
			},
		})
	} else {
		c.JSON(http.StatusOK, models.SMSResponse{
			Success: false,
			Message: "SMS failed to send",
			Error:   errorMsg,
			Data: map[string]interface{}{
				"log_id":    smsLog.ID,
				"recipient": smsLog.Recipient,
				"status":    status,
			},
		})
	}
}

// SendBulkSMS handles sending multiple SMS messages
func (h *SMSHandler) SendBulkSMS(c *gin.Context) {
	var req models.BulkSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.SMSResponse{
			Success: false,
			Message: "Invalid request payload",
			Error:   err.Error(),
		})
		return
	}

	// Validate all phone numbers
	for _, msg := range req.Messages {
		if !utils.ValidatePhone(msg.Number) {
			c.JSON(http.StatusBadRequest, models.SMSResponse{
				Success: false,
				Message: "Invalid phone number in messages",
				Error:   "Phone number " + msg.Number + " is invalid",
			})
			return
		}
	}

	// Get client from context
	client, exists := c.Get("client")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Client not found in context",
		})
		return
	}
	apiClient := client.(models.APIClient)
	clientID := apiClient.ID

	// Check if bulk request exceeds limits
	if apiClient.DailyUsage+len(req.Messages) > apiClient.DailyLimit {
		c.JSON(http.StatusTooManyRequests, models.SMSResponse{
			Success: false,
			Message: "Bulk request would exceed daily limit",
			Error:   "Requested messages exceed available daily quota",
		})
		return
	}

	// Send SMS via provider
	responses, err := h.smsProvider.SendSMS(req.Messages, apiClient.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to send bulk SMS",
			Error:   err.Error(),
		})
		return
	}

	// Log all SMS messages and track results
	results := make([]map[string]interface{}, 0)
	successCount := 0
	failedCount := 0

	for i, msg := range req.Messages {
		var resp service.SMSProviderResponse
		if i < len(responses) {
			resp = responses[i]
		} else if len(responses) > 0 {
			// Use first response if not enough responses
			resp = responses[0]
		}

		status := "sent"
		providerStatus := "Success"
		providerMessage := "SMS sent successfully"
		errorMsg := ""

		if resp.Status != "Success" && resp.Status != "success" {
			status = "failed"
			providerStatus = resp.Status
			providerMessage = resp.Message
			errorMsg = resp.Message
			failedCount++
		} else {
			successCount++
		}

		// Log SMS
		smsLog := models.SMSLog{
			ClientID:       clientID,
			Recipient:      utils.FormatPhone(msg.Number),
			Message:        msg.Message,
			SenderID:       msg.SenderID,
			Priority:       msg.Priority,
			Status:         status,
			ProviderStatus: providerStatus,
			ProviderMessage: providerMessage,
			Error:          errorMsg,
			IPAddress:      c.ClientIP(),
			UserAgent:      c.GetHeader("User-Agent"),
		}
		database.DB.Create(&smsLog)

		results = append(results, map[string]interface{}{
			"log_id":    smsLog.ID,
			"recipient": smsLog.Recipient,
			"status":    status,
		})
	}

	// Update client usage
	apiClient.DailyUsage += successCount
	apiClient.MonthlyUsage += successCount
	database.DB.Save(&apiClient)

	c.JSON(http.StatusOK, models.SMSResponse{
		Success: true,
		Message: "Bulk SMS processing completed",
		Data: map[string]interface{}{
			"total":        len(req.Messages),
			"successful":   successCount,
			"failed":       failedCount,
			"results":      results,
		},
	})
}

// GetSMSLogs retrieves SMS logs for the authenticated client
func (h *SMSHandler) GetSMSLogs(c *gin.Context) {
	clientID, exists := c.Get("client_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Client ID not found",
		})
		return
	}

	var logs []models.SMSLog
	query := database.DB.Where("client_id = ?", clientID).Order("created_at DESC")

	// Pagination
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")
	query = query.Limit(limit).Offset(offset)

	// Status filter
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Failed to retrieve logs",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.SMSResponse{
		Success: true,
		Message: "Logs retrieved successfully",
		Data:    logs,
	})
}

// GetStats returns usage statistics for the authenticated client
func (h *SMSHandler) GetStats(c *gin.Context) {
	client, exists := c.Get("client")
	if !exists {
		c.JSON(http.StatusInternalServerError, models.SMSResponse{
			Success: false,
			Message: "Client not found",
		})
		return
	}
	apiClient := client.(models.APIClient)

	stats := models.ClientStats{
		ClientID:     apiClient.ID,
		DailyUsage:   apiClient.DailyUsage,
		MonthlyUsage: apiClient.MonthlyUsage,
		DailyLimit:   apiClient.DailyLimit,
		MonthlyLimit: apiClient.MonthlyLimit,
		IsActive:     apiClient.IsActive,
	}

	c.JSON(http.StatusOK, models.SMSResponse{
		Success: true,
		Message: "Statistics retrieved successfully",
		Data:    stats,
	})
}

