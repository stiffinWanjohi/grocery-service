package notification

import (
	"context"

	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/utils"
)

type NotificationService interface {
	SendOrderConfirmation(ctx context.Context, order *domain.Order) error
	SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error
}

type CompositeNotificationService struct {
	services []NotificationService
}

func NewCompositeNotificationService(services ...NotificationService) NotificationService {
	return &CompositeNotificationService{services: services}
}

func (s *CompositeNotificationService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	var lastError error
	for _, service := range s.services {
		if err := service.SendOrderConfirmation(ctx, order); err != nil {
			lastError = utils.LogError(err,
				"Failed to send order confirmation",
				utils.ErrCodeEmailSendFailed).Error
		}
	}
	return lastError
}

func (s *CompositeNotificationService) SendOrderStatusUpdate(ctx context.Context, order *domain.Order) error {
	var lastError error
	for _, service := range s.services {
		if err := service.SendOrderStatusUpdate(ctx, order); err != nil {
			lastError = utils.LogError(err,
				"Failed to send status update",
				utils.ErrCodeEmailSendFailed).Error
		}
	}
	return lastError
}
