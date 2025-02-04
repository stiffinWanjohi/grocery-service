package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/utils"
	at "github.com/kamikazechaser/africastalking/v2"
)

type SMSService struct {
	config config.SMSConfig
	client *at.AfricasTalking
}

func NewSMSService(config config.SMSConfig) (*SMSService, error) {
	client := at.Initialize(config.Username, config.APIKey)
	if client == nil {
		return nil, utils.WrapError(fmt.Errorf("client initialization failed"), "failed to initialize SMS service")
	}

	if config.Environment == "sandbox" {
		client.SetEnvironment(at.Sandbox)
	} else {
		client.SetEnvironment(at.Production)
	}

	return &SMSService{
		config: config,
		client: client,
	}, nil
}

func (s *SMSService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	message := fmt.Sprintf("Order #%s confirmed. Total: %.2f. Thank you for your order!",
		order.ID, order.TotalPrice)
	if err := s.sendSMS(ctx, order.Customer.Phone, message); err != nil {
		return utils.LogError(err, "Failed to send order confirmation SMS", utils.ErrCodeSMSSendFailed).Error
	}
	return nil
}

func (s *SMSService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error {
	message := fmt.Sprintf("Order #%s status updated to: %s",
		order.ID, order.Status)
	if err := s.sendSMS(ctx, order.Customer.Phone, message); err != nil {
		return utils.LogError(err, "Failed to send order status update SMS", utils.ErrCodeSMSSendFailed).Error
	}
	return nil
}

func (s *SMSService) sendSMS(ctx context.Context, phone, message string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		response, err := s.client.SMS().Send(&at.SendSMSInput{
			To:      []string{phone},
			Message: message,
			From:    s.config.SenderID,
		})

		if err != nil {
			done <- utils.WrapError(err, "failed to send SMS")
			return
		}

		if len(response.Recipients) == 0 {
			done <- fmt.Errorf("SMS not delivered: no recipients")
			return
		}

		recipient := response.Recipients[0]
		if recipient.Status != "Success" {
			done <- fmt.Errorf("SMS delivery failed: %s", recipient.StatusReason)
			return
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		return utils.WrapError(ctx.Err(), "SMS sending timeout")
	case err := <-done:
		return err
	}
}
