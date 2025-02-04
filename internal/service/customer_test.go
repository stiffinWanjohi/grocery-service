package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	mocks "github.com/grocery-service/tests/mocks/repository"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
)

func TestCustomerService_Create(t *testing.T) {
	mockRepo := mocks.NewCustomerRepository(t)
	service := NewCustomerService(mockRepo)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
	}

	// Test case: Successful creation
	mockRepo.On("GetByEmail", ctx, customer.Email).Return(nil, customErrors.ErrCustomerNotFound)
	mockRepo.On("Create", ctx, customer).Return(nil)

	err := service.Create(ctx, customer)
	assert.NoError(t, err)

	// Test case: Email already registered
	existingCustomer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "Existing User",
		Email: "john@example.com",
	}
	mockRepo.On("GetByEmail", ctx, existingCustomer.Email).Return(existingCustomer, nil)

	err = service.Create(ctx, existingCustomer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already registered")

	// Test validation errors
	testCases := []struct {
		name     string
		customer *domain.Customer
		errMsg   string
	}{
		{
			name:     "Empty name",
			customer: &domain.Customer{ID: uuid.New(), Email: "john@example.com", Phone: "+1234567890"},
			errMsg:   "customer name is required",
		},
		{
			name:     "Invalid email",
			customer: &domain.Customer{ID: uuid.New(), Name: "John Doe", Email: "invalid-email", Phone: "+1234567890"},
			errMsg:   "invalid email format",
		},
		{
			name:     "Invalid phone",
			customer: &domain.Customer{ID: uuid.New(), Name: "John Doe", Email: "john@example.com", Phone: "invalid"},
			errMsg:   "invalid phone format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo.On("GetByEmail", ctx, tc.customer.Email).Return(nil, customErrors.ErrCustomerNotFound)
			err := service.Create(ctx, tc.customer)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestCustomerService_GetByID(t *testing.T) {
	mockRepo := mocks.NewCustomerRepository(t)
	service := NewCustomerService(mockRepo)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
	}

	// Test empty ID
	_, err := service.GetByID(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "customer ID is required")

	// Test successful retrieval
	mockRepo.On("GetByID", ctx, customer.ID.String()).Return(customer, nil)
	found, err := service.GetByID(ctx, customer.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, found.ID)

	// Test not found case
	mockRepo.On("GetByID", ctx, "non-existent").Return(nil, customErrors.ErrCustomerNotFound)
	_, err = service.GetByID(ctx, "non-existent")
	assert.ErrorIs(t, err, customErrors.ErrCustomerNotFound)
}

func TestCustomerService_GetByEmail(t *testing.T) {
	mockRepo := mocks.NewCustomerRepository(t)
	service := NewCustomerService(mockRepo)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
	}

	// Test empty email
	_, err := service.GetByEmail(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")

	// Test successful retrieval
	mockRepo.On("GetByEmail", ctx, customer.Email).Return(customer, nil)
	found, err := service.GetByEmail(ctx, customer.Email)
	assert.NoError(t, err)
	assert.Equal(t, customer.ID, found.ID)

	// Test not found case
	mockRepo.On("GetByEmail", ctx, "nonexistent@example.com").Return(nil, customErrors.ErrCustomerNotFound)
	_, err = service.GetByEmail(ctx, "nonexistent@example.com")
	assert.ErrorIs(t, err, customErrors.ErrCustomerNotFound)
}

func TestCustomerService_List(t *testing.T) {
	mockRepo := mocks.NewCustomerRepository(t)
	service := NewCustomerService(mockRepo)
	ctx := context.Background()

	customers := []domain.Customer{
		{ID: uuid.New(), Name: "John Doe", Email: "john@example.com", Phone: "+1234567890"},
		{ID: uuid.New(), Name: "Jane Doe", Email: "jane@example.com", Phone: "+1234567891"},
	}

	mockRepo.On("List", ctx).Return(customers, nil)

	found, err := service.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, len(customers))
}

func TestCustomerService_Update(t *testing.T) {
	mockRepo := mocks.NewCustomerRepository(t)
	service := NewCustomerService(mockRepo)
	ctx := context.Background()

	customer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "+1234567890",
	}

	// Test successful update
	mockRepo.On("GetByEmail", ctx, customer.Email).Return(nil, customErrors.ErrCustomerNotFound)
	mockRepo.On("Update", ctx, customer).Return(nil)

	err := service.Update(ctx, customer)
	assert.NoError(t, err)

	// Test email already registered by another customer
	existingCustomer := &domain.Customer{
		ID:    uuid.New(),
		Name:  "Another User",
		Email: "john@example.com",
	}
	mockRepo.On("GetByEmail", ctx, existingCustomer.Email).Return(existingCustomer, nil)

	err = service.Update(ctx, existingCustomer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already registered")

	// Test validation error
	invalidCustomer := &domain.Customer{
		ID: uuid.New(),
	}
	mockRepo.On("GetByEmail", ctx, invalidCustomer.Email).Return(nil, customErrors.ErrCustomerNotFound)

	err = service.Update(ctx, invalidCustomer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "customer name is required")
}

func TestCustomerService_Delete(t *testing.T) {
	mockRepo := mocks.NewCustomerRepository(t)
	service := NewCustomerService(mockRepo)
	ctx := context.Background()

	id := uuid.New().String()

	// Test empty ID
	err := service.Delete(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "customer ID is required")

	// Test customer not found
	mockRepo.On("GetByID", ctx, "non-existent").Return(nil, customErrors.ErrCustomerNotFound)
	err = service.Delete(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find customer")

	// Test successful deletion
	existingCustomer := &domain.Customer{
		ID:    uuid.MustParse(id),
		Name:  "John Doe",
		Email: "john@example.com",
	}
	mockRepo.On("GetByID", ctx, id).Return(existingCustomer, nil)
	mockRepo.On("Delete", ctx, id).Return(nil)

	err = service.Delete(ctx, id)
	assert.NoError(t, err)
}
