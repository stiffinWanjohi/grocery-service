// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/grocery-service/internal/domain"
	mock "github.com/stretchr/testify/mock"
)

// OrderService is an autogenerated mock type for the OrderService type
type OrderService struct {
	mock.Mock
}

// AddOrderItem provides a mock function with given fields: ctx, orderID, item
func (_m *OrderService) AddOrderItem(ctx context.Context, orderID string, item *domain.OrderItem) error {
	ret := _m.Called(ctx, orderID, item)

	if len(ret) == 0 {
		panic("no return value specified for AddOrderItem")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *domain.OrderItem) error); ok {
		r0 = rf(ctx, orderID, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: ctx, order
func (_m *OrderService) Create(ctx context.Context, order *domain.Order) error {
	ret := _m.Called(ctx, order)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Order) error); ok {
		r0 = rf(ctx, order)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *OrderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetByID")
	}

	var r0 *domain.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*domain.Order, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *domain.Order); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx
func (_m *OrderService) List(ctx context.Context) ([]domain.Order, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []domain.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]domain.Order, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []domain.Order); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]domain.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByCustomerID provides a mock function with given fields: ctx, customerID
func (_m *OrderService) ListByCustomerID(ctx context.Context, customerID string) ([]domain.Order, error) {
	ret := _m.Called(ctx, customerID)

	if len(ret) == 0 {
		panic("no return value specified for ListByCustomerID")
	}

	var r0 []domain.Order
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]domain.Order, error)); ok {
		return rf(ctx, customerID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []domain.Order); ok {
		r0 = rf(ctx, customerID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]domain.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, customerID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveOrderItem provides a mock function with given fields: ctx, orderID, itemID
func (_m *OrderService) RemoveOrderItem(ctx context.Context, orderID string, itemID string) error {
	ret := _m.Called(ctx, orderID, itemID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveOrderItem")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, orderID, itemID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, order
func (_m *OrderService) Update(ctx context.Context, order *domain.Order) error {
	ret := _m.Called(ctx, order)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Order) error); ok {
		r0 = rf(ctx, order)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateStatus provides a mock function with given fields: ctx, id, status
func (_m *OrderService) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	ret := _m.Called(ctx, id, status)

	if len(ret) == 0 {
		panic("no return value specified for UpdateStatus")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, domain.OrderStatus) error); ok {
		r0 = rf(ctx, id, status)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewOrderService creates a new instance of OrderService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOrderService(t interface {
	mock.TestingT
	Cleanup(func())
}) *OrderService {
	mock := &OrderService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
