package notification

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/grocery-service/internal/domain"
)

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

type EmailService struct {
	config EmailConfig
}

func NewEmailService(config EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

func (s *EmailService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	subject := fmt.Sprintf("Order Confirmation #%s", order.ID)
	body := s.generateOrderConfirmationEmail(order)
	return s.SendEmail(ctx, order.Customer.Email, subject, body)
}

func (s *EmailService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error {
	subject := fmt.Sprintf("Order Status Update #%s", order.ID)
	body := s.generateOrderStatusUpdateEmail(order)
	return s.SendEmail(ctx, order.Customer.Email, subject, body)
}

func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

		msg := fmt.Sprintf("From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n", s.config.FromName, s.config.FromEmail, to, subject, body)

		addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
		done <- smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(msg))
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email sending timeout: %w", ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	}
}

func (s *EmailService) generateOrderConfirmationEmail(order *domain.Order) string {
	template := `
		<h2>Order Confirmation</h2>
		<p>Dear %s,</p>
		<p>Thank you for your order. Here are your order details:</p>
		<p><strong>Order ID:</strong> %s</p>
		<p><strong>Total Amount:</strong> %.2f</p>
		<h3>Order Items:</h3>
		<ul>
	`
	itemsList := ""
	for _, item := range order.Items {
		itemsList += fmt.Sprintf("<li>%s - Quantity: %d - Price: %.2f</li>",
			item.Product.Name, item.Quantity, item.Price)
	}

	footer := `
		</ul>
		<p>We will notify you when your order status changes.</p>
		<p>Best regards,<br>Grocery Service Team</p>
	`

	return fmt.Sprintf(template, order.Customer.Name, order.ID, order.TotalPrice) + itemsList + footer
}

func (s *EmailService) generateOrderStatusUpdateEmail(order *domain.Order) string {
	template := `
		<h2>Order Status Update</h2>
		<p>Dear %s,</p>
		<p>Your order status has been updated:</p>
		<p><strong>Order ID:</strong> %s</p>
		<p><strong>New Status:</strong> %s</p>
		<p>If you have any questions, please contact our support team.</p>
		<p>Best regards,<br>Grocery Service Team</p>
	`
	return fmt.Sprintf(template, order.Customer.Name, order.ID, order.Status)
}
