package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/utils/errors"
)

type SMSService struct {
	config config.SMSConfig
	client *http.Client
}

func NewSMSService(config config.SMSConfig) *SMSService {
	return &SMSService{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SMSService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	message := fmt.Sprintf(
		"Hi %s, Order #%s confirmed. Total: %.2f. Delivery to: %s. Thank you for your order!",
		order.Customer.User.Name,
		order.ID,
		order.TotalPrice,
		order.Customer.User.Address,
	)
	return s.sendSMS(ctx, order.Customer.User.Phone, message)
}

func (s *SMSService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error {
	message := fmt.Sprintf(
		"Hi %s, Order #%s status updated to: %s. Delivery to: %s",
		order.Customer.User.Name,
		order.ID,
		order.Status,
		order.Customer.User.Address,
	)
	return s.sendSMS(ctx, order.Customer.User.Phone, message)
}

func (s *SMSService) sendSMS(ctx context.Context, phone, message string) error {
	payload := map[string]string{
		"username": s.config.Username,
		"to":       phone,
		"message":  message,
		"from":     s.config.SenderID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.WrapError(err, "failed to marshal SMS payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.BaseURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return errors.WrapError(err, "failed to create SMS request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return errors.WrapError(err, "failed to send SMS request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send SMS: received status %s", resp.Status)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return errors.WrapError(err, "failed to decode SMS response")
	}

	smsData, ok := response["SMSMessageData"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid SMS response structure")
	}

	recipients, ok := smsData["Recipients"].([]interface{})
	if !ok || len(recipients) == 0 {
		return fmt.Errorf("SMS not delivered: no recipients")
	}

	firstRecipient := recipients[0].(map[string]interface{})
	if firstRecipient["status"] != "Success" {
		return fmt.Errorf("SMS delivery failed: %s", firstRecipient["status"])
	}

	return nil
}
