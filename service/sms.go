package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Ian-Balijawa/sms-gateway/config"
	"github.com/Ian-Balijawa/sms-gateway/models"
	"github.com/Ian-Balijawa/sms-gateway/utils"
)

type SMSProviderResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

type SMSProvider struct {
	client *http.Client
}

func NewSMSProvider() *SMSProvider {
	return &SMSProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SMSProvider) GetAPIURL() string {
	if config.AppConfig.SMSSandboxMode {
		return config.AppConfig.SMSSandboxURL
	}
	return config.AppConfig.SMSLiveURL
}

func (s *SMSProvider) SendSMS(messages []models.SMSRequest, defaultSenderID string) ([]SMSProviderResponse, error) {
	// Prepare payload matching the egosms.co API format
	payload := map[string]interface{}{
		"method": "SendSms",
		"userdata": map[string]string{
			"username": config.AppConfig.SMSUsername,
			"password": config.AppConfig.SMSPassword,
		},
		"msgdata": make([]map[string]interface{}, 0),
	}

	// Map messages to API format
	for _, msg := range messages {
		senderID := msg.SenderID
		if senderID == "" {
			senderID = defaultSenderID
		}
		priority := msg.Priority
		if priority == "" {
			priority = "1"
		}

		payload["msgdata"] = append(payload["msgdata"].([]map[string]interface{}), map[string]interface{}{
			"number":   utils.FormatPhone(msg.Number),
			"message":  msg.Message,
			"senderid": senderID,
			"priority": priority,
		})
	}

	// Serialize payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", s.GetAPIURL(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMS request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var providerResp SMSProviderResponse
	if err := json.Unmarshal(body, &providerResp); err != nil {
		// If response is not in expected format, return raw response
		log.Printf("Unexpected response format: %s", string(body))
		return []SMSProviderResponse{{
			Status:  "Failed",
			Message: string(body),
		}}, nil
	}

	// Return array of responses (provider may return single or multiple)
	responses := []SMSProviderResponse{providerResp}

	// Log the response
	log.Printf("SMS Provider Response: Status=%s, Message=%s", providerResp.Status, providerResp.Message)

	return responses, nil
}

