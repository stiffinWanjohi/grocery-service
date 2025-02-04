package notification

import (
	"context"

	"github.com/grocery-service/internal/domain"
)

type NotificationService interface {
	SendOrderConfirmation(ctx context.Context, order *domain.Order) error
	SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error
}

// CompositeNotificationService allows sending notifications through multiple channels
type CompositeNotificationService struct {
	services []NotificationService
}

func NewCompositeNotificationService(services ...NotificationService) NotificationService {
	return &CompositeNotificationService{services: services}
}

func (s *CompositeNotificationService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	for _, service := range s.services {
		if err := service.SendOrderConfirmation(ctx, order); err != nil {
			// Log error but continue with other services
			continue
		}
	}
	return nil
}

func (s *CompositeNotificationService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error {
	for _, service := range s.services {
		if err := service.SendOrderStatusUpdate(ctx, order); err != nil {
			// Log error but continue with other services
			continue
		}
	}
	return nil
}
