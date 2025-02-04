package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/config"
	"github.com/grocery-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestSMSService_SendOrderConfirmation(t *testing.T) {
	// Mock SMS API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("apiKey"))

		response := map[string]interface{}{
			"SMSMessageData": map[string]interface{}{
				"Recipients": []interface{}{
					map[string]interface{}{
						"status": "Success",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := config.SMSConfig{
		BaseURL:  server.URL,
		Username: "testuser",
		APIKey:   "test-api-key",
		SenderID: "TEST",
	}

	service := NewSMSService(config)
	order := &domain.Order{
		ID:         uuid.New(),
		TotalPrice: 100.50,
		Customer: &domain.Customer{
			Phone: "+1234567890",
		},
	}

	err := service.SendOrderConfirmation(context.Background(), order)
	assert.NoError(t, err)
}

func TestEmailService_SendOrderConfirmation(t *testing.T) {
	// Create mock SMTP server
	mockSMTP := &mockSMTPServer{t: t}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer listener.Close()

	go mockSMTP.Start(listener)

	config := config.SMTPConfig{
		Host:     "127.0.0.1",
		Port:     listener.Addr().(*net.TCPAddr).Port,
		Username: "test@test.com",
		Password: "password",
		From:     "noreply@test.com",
		FromName: "Test Service",
	}

	service := NewEmailService(config)
	order := &domain.Order{
		ID:         uuid.New(),
		TotalPrice: 100.50,
		Customer: &domain.Customer{
			Name:  "John Doe",
			Email: "john@example.com",
		},
		Items: []domain.OrderItem{
			{
				Product:  &domain.Product{Name: "Test Product"},
				Quantity: 2,
				Price:    50.25,
			},
		},
	}

	err = service.SendOrderConfirmation(context.Background(), order)
	assert.NoError(t, err)

	// Verify email content
	assert.True(t, strings.Contains(mockSMTP.LastMessage(), "Order Confirmation"))
	assert.True(t, strings.Contains(mockSMTP.LastMessage(), "John Doe"))
	assert.True(t, strings.Contains(mockSMTP.LastMessage(), "Test Product"))
}

// Mock SMTP server
type mockSMTPServer struct {
	t       *testing.T
	lastMsg string
}

func (s *mockSMTPServer) Start(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *mockSMTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Send greeting
	conn.Write([]byte("220 mock.smtp.server\r\n"))

	buf := make([]byte, 1024)
	var message strings.Builder

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		cmd := string(buf[:n])
		message.WriteString(cmd)

		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			conn.Write([]byte("250-mock.smtp.server\r\n250 AUTH LOGIN PLAIN\r\n"))
		case strings.HasPrefix(cmd, "AUTH"):
			conn.Write([]byte("235 Authentication successful\r\n"))
		case strings.HasPrefix(cmd, "MAIL FROM"):
			conn.Write([]byte("250 Sender OK\r\n"))
		case strings.HasPrefix(cmd, "RCPT TO"):
			conn.Write([]byte("250 Recipient OK\r\n"))
		case strings.HasPrefix(cmd, "DATA"):
			conn.Write([]byte("354 Enter message\r\n"))
		case strings.Contains(cmd, "\r\n.\r\n"):
			conn.Write([]byte("250 Message received\r\n"))
			s.lastMsg = message.String()
			return
		case strings.HasPrefix(cmd, "QUIT"):
			conn.Write([]byte("221 Goodbye\r\n"))
			return
		}
	}
}

func (s *mockSMTPServer) LastMessage() string {
	return s.lastMsg
}

func TestCompositeNotificationService(t *testing.T) {
	mockSMS := &mockNotificationService{}
	mockEmail := &mockNotificationService{}

	service := NewCompositeNotificationService(mockSMS, mockEmail)
	order := &domain.Order{ID: uuid.New()}

	// Test successful notifications
	err := service.SendOrderConfirmation(context.Background(), order)
	assert.NoError(t, err)

	// Test partial failure
	mockSMS.shouldFail = true
	err = service.SendOrderConfirmation(context.Background(), order)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to send order confirmation")
}

// Mock notification service for testing
type mockNotificationService struct {
	shouldFail bool
}

func (m *mockNotificationService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}
	return nil
}

func (m *mockNotificationService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error {
	if m.shouldFail {
		return fmt.Errorf("mock error")
	}
	return nil
}
