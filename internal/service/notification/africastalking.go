package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/africastalking/africastalking-go/sms"
	"github.com/grocery-service/internal/domain"
)

type AfricasTalkingConfig struct {
	APIKey      string
	Username    string
	SenderID    string
	Environment string
}

type AfricasTalkingService struct {
	config AfricasTalkingConfig
	client *sms.Service
}

func NewAfricasTalkingService(config AfricasTalkingConfig) (*AfricasTalkingService, error) {
	client, err := sms.NewService(sms.Options{
		ApiKey:    config.APIKey,
		Username:  config.Username,
		IsSandbox: config.Environment == "sandbox",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Africa's Talking service: %w", err)
	}

	return &AfricasTalkingService{
		config: config,
		client: client,
	}, nil
}

func (s *AfricasTalkingService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	message := fmt.Sprintf("Thank you for your order #%s. Total amount: %.2f. We'll update you on the status.",
		order.ID, order.TotalPrice)
	return s.SendSMS(ctx, order.Customer.Phone, message)
}

func (s *AfricasTalkingService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error {
	message := fmt.Sprintf("Order #%s status update: %s. For support, contact our team.",
		order.ID, order.Status)
	return s.SendSMS(ctx, order.Customer.Phone, message)
}

func (s *AfricasTalkingService) SendSMS(ctx context.Context, phone, message string) error {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		response, err := s.client.Send(sms.SendMessageRequest{
			To:      []string{phone},
			Message: message,
			From:    s.config.SenderID,
		})

		if err != nil {
			done <- fmt.Errorf("failed to send SMS: %w", err)
			return
		}

		if len(response.SMSMessageData.Recipients) == 0 {
			done <- fmt.Errorf("SMS not delivered: no recipients")
			return
		}

		recipient := response.SMSMessageData.Recipients[0]
		if recipient.Status != "Success" {
			done <- fmt.Errorf("SMS delivery failed: %s", recipient.StatusReason)
			return
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("SMS sending timeout: %w", ctx.Err())
	case err := <-done:
		return err
	}
}
