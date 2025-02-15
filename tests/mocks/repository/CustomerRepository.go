// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/grocery-service/internal/domain"
	mock "github.com/stretchr/testify/mock"
)

// CustomerRepository is an autogenerated mock type for the CustomerRepository type
type CustomerRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, customer
func (_m *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	ret := _m.Called(ctx, customer)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Customer) error); ok {
		r0 = rf(ctx, customer)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, id
func (_m *CustomerRepository) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *CustomerRepository) GetByID(ctx context.Context, id string) (*domain.Customer, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for GetByID")
	}

	var r0 *domain.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*domain.Customer, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *domain.Customer); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Customer)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByUserID provides a mock function with given fields: ctx, userID
func (_m *CustomerRepository) GetByUserID(ctx context.Context, userID string) (*domain.Customer, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetByUserID")
	}

	var r0 *domain.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*domain.Customer, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *domain.Customer); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Customer)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx
func (_m *CustomerRepository) List(ctx context.Context) ([]domain.Customer, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []domain.Customer
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]domain.Customer, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []domain.Customer); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]domain.Customer)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, customer
func (_m *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	ret := _m.Called(ctx, customer)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Customer) error); ok {
		r0 = rf(ctx, customer)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewCustomerRepository creates a new instance of CustomerRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCustomerRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *CustomerRepository {
	mock := &CustomerRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
