package notification

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
)

type EmailService struct {
	config config.SMTPConfig
}

func NewEmailService(
	config config.SMTPConfig,
) *EmailService {
	return &EmailService{
		config: config,
	}
}

func (s *EmailService) SendOrderConfirmation(
	ctx context.Context,
	order *domain.Order,
) error {
	subject := fmt.Sprintf(
		"Order Confirmation #%s",
		order.ID,
	)

	body := s.generateOrderConfirmationEmail(order)

	return s.SendEmail(
		ctx,
		order.Customer.User.Email,
		subject,
		body,
	)
}

func (s *EmailService) SendOrderStatusUpdate(
	ctx context.Context,
	order *domain.Order,
) error {
	subject := fmt.Sprintf(
		"Order Status Update #%s",
		order.ID,
	)

	body := s.generateOrderStatusUpdateEmail(order)

	return s.SendEmail(
		ctx,
		order.Customer.User.Email,
		subject,
		body,
	)
}

func (s *EmailService) SendEmail(
	ctx context.Context,
	to, subject, body string,
) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		auth := smtp.PlainAuth(
			"",
			s.config.Username,
			s.config.Password,
			s.config.Host,
		)
		msg := fmt.Sprintf("From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n", s.config.FromName, s.config.From, to, subject, body)
		addr := fmt.Sprintf(
			"%s:%d",
			s.config.Host,
			s.config.Port,
		)
		done <- smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf(
			"email sending timeout: %w",
			ctx.Err(),
		)
	case err := <-done:
		if err != nil {
			return fmt.Errorf(
				"failed to send email: %w",
				err,
			)
		}
		return nil
	}
}

func (s *EmailService) generateOrderConfirmationEmail(
	order *domain.Order,
) string {
	template := `
		<h2>Order Confirmation</h2>
		<p>Dear %s,</p>
		<p>Thank you for your order. Here are your order details:</p>
		<p><strong>Order ID:</strong> %s</p>
		<p><strong>Total Amount:</strong> %.2f</p>
		<p><strong>Delivery Address:</strong> %s</p>
		<p><strong>Contact Phone:</strong> %s</p>
		<h3>Order Items:</h3>
		<ul>
	`
	itemsList := ""

	for _, item := range order.Items {
		itemsList += fmt.Sprintf(
			"<li>%s - Quantity: %d - Price: %.2f</li>",
			item.Product.Name,
			item.Quantity,
			item.Price,
		)
	}

	footer := `
		</ul>
		<p>We will notify you when your order status changes.</p>
		<p>Best regards,<br>Grocery Service Team</p>
	`

	return fmt.Sprintf(
		template,
		order.Customer.User.Name,
		order.ID,
		order.TotalPrice,
		order.Customer.User.Address,
		order.Customer.User.Phone,
	) + itemsList + footer
}

func (s *EmailService) generateOrderStatusUpdateEmail(
	order *domain.Order,
) string {
	template := `
		<h2>Order Status Update</h2>
		<p>Dear %s,</p>
		<p>Your order status has been updated:</p>
		<p><strong>Order ID:</strong> %s</p>
		<p><strong>New Status:</strong> %s</p>
		<p><strong>Delivery Address:</strong> %s</p>
		<p><strong>Contact Phone:</strong> %s</p>
		<p>If you have any questions, please contact our support team.</p>
		<p>Best regards,<br>Grocery Service Team</p>
	`
	return fmt.Sprintf(template,
		order.Customer.User.Name,
		order.ID,
		order.Status,
		order.Customer.User.Address,
		order.Customer.User.Phone,
	)
}
