// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	db "github.com/grocery-service/internal/repository/db"
	gorm "gorm.io/gorm"

	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository[T interface{}] struct {
	mock.Mock
}

// WithTx provides a mock function with given fields: tx
func (_m *Repository[T]) WithTx(tx *gorm.DB) db.Repository[T] {
	ret := _m.Called(tx)

	if len(ret) == 0 {
		panic("no return value specified for WithTx")
	}

	var r0 db.Repository[T]
	if rf, ok := ret.Get(0).(func(*gorm.DB) db.Repository[T]); ok {
		r0 = rf(tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(db.Repository[T])
		}
	}

	return r0
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepository[T interface{}](t interface {
	mock.TestingT
	Cleanup(func())
}) *Repository[T] {
	mock := &Repository[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
