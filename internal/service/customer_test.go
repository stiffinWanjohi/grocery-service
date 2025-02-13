package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	mocks "github.com/grocery-service/tests/mocks/repository"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCustomerTest(
	t *testing.T,
) (
	CustomerService,
	*mocks.CustomerRepository,
	*mocks.UserRepository,
) {
	customerRepo := mocks.NewCustomerRepository(t)
	userRepo := mocks.NewUserRepository(t)
	service := NewCustomerService(customerRepo, userRepo)
	return service, customerRepo, userRepo
}

func createTestUser() *domain.User {
	return &domain.User{
		ID:      uuid.New(),
		Email:   "test@example.com",
		Name:    "Test User",
		Role:    domain.CustomerRole,
		Picture: "https://example.com/picture.jpg",
	}
}

func TestCustomerService_Create(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*mocks.CustomerRepository, *mocks.UserRepository, string)
		expectedError error
	}{
		{
			name:   "Success - Create New Customer",
			userID: "google-oauth2|123456789",
			setupMocks: func(cr *mocks.CustomerRepository, ur *mocks.UserRepository, userID string) {
				user := createTestUser()
				ur.On("GetByProviderID", mock.Anything, userID).Return(user, nil)
				cr.On("GetByUserID", mock.Anything, user.ID.String()).
					Return(nil, customErrors.ErrCustomerNotFound)
				cr.On(
					"Create",
					mock.Anything,
					mock.MatchedBy(func(c *domain.Customer) bool {
						return c.UserID == user.ID
					})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Error - Empty UserID",
			userID:        "",
			setupMocks:    func(_ *mocks.CustomerRepository, _ *mocks.UserRepository, _ string) {},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name:   "Error - User Not Found",
			userID: "google-oauth2|nonexistent",
			setupMocks: func(_ *mocks.CustomerRepository, ur *mocks.UserRepository, userID string) {
				ur.On("GetByProviderID", mock.Anything, userID).
					Return(nil, customErrors.ErrUserNotFound)
			},
			expectedError: customErrors.ErrUserNotFound,
		},
		{
			name:   "Error - Customer Already Exists",
			userID: "google-oauth2|existing",
			setupMocks: func(cr *mocks.CustomerRepository, ur *mocks.UserRepository, userID string) {
				user := createTestUser()
				ur.On("GetByProviderID", mock.Anything, userID).Return(user, nil)
				cr.On("GetByUserID", mock.Anything, user.ID.String()).
					Return(&domain.Customer{
						ID:     uuid.New(),
						UserID: user.ID,
						User:   user,
					}, nil)
			},
			expectedError: customErrors.ErrCustomerExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, customerRepo, userRepo := setupCustomerTest(t)
			tt.setupMocks(customerRepo, userRepo, tt.userID)

			customer, err := service.Create(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, customer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, customer)
				assert.NotEmpty(t, customer.ID)
			}
		})
	}
}

func TestCustomerService_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		customerID    string
		setupMocks    func(*mocks.CustomerRepository, string)
		expectedError error
	}{
		{
			name:       "Success - Get Existing Customer",
			customerID: uuid.New().String(),
			setupMocks: func(cr *mocks.CustomerRepository, customerID string) {
				user := createTestUser()
				customer := &domain.Customer{
					ID:     uuid.MustParse(customerID),
					UserID: user.ID,
					User:   user,
				}
				cr.On("GetByID", mock.Anything, customerID).
					Return(customer, nil)
			},
			expectedError: nil,
		},
		{
			name:          "Error - Empty CustomerID",
			customerID:    "",
			setupMocks:    func(_ *mocks.CustomerRepository, _ string) {},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name:       "Error - Customer Not Found",
			customerID: uuid.New().String(),
			setupMocks: func(cr *mocks.CustomerRepository, customerID string) {
				cr.On("GetByID", mock.Anything, customerID).
					Return(nil, customErrors.ErrCustomerNotFound)
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, customerRepo, _ := setupCustomerTest(t)
			tt.setupMocks(customerRepo, tt.customerID)

			customer, err := service.GetByID(
				context.Background(),
				tt.customerID,
			)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, customer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, customer)
				assert.Equal(t, tt.customerID, customer.ID.String())
				assert.NotNil(t, customer.User)
			}
		})
	}
}

func TestCustomerService_GetByUserID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*mocks.CustomerRepository, string)
		expectedError error
	}{
		{
			name:   "Success - Get Customer By UserID",
			userID: uuid.New().String(),
			setupMocks: func(cr *mocks.CustomerRepository, userID string) {
				user := createTestUser()
				user.ID = uuid.MustParse(userID)
				customer := &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
					User:   user,
				}
				cr.On("GetByUserID", mock.Anything, userID).
					Return(customer, nil)
			},
			expectedError: nil,
		},
		{
			name:          "Error - Empty UserID",
			userID:        "",
			setupMocks:    func(_ *mocks.CustomerRepository, _ string) {},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name:   "Error - Customer Not Found",
			userID: uuid.New().String(),
			setupMocks: func(cr *mocks.CustomerRepository, userID string) {
				cr.On("GetByUserID", mock.Anything, userID).
					Return(nil, customErrors.ErrCustomerNotFound)
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, customerRepo, _ := setupCustomerTest(t)
			tt.setupMocks(customerRepo, tt.userID)

			customer, err := service.GetByUserID(
				context.Background(),
				tt.userID,
			)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, customer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, customer)
				assert.Equal(t, tt.userID, customer.UserID.String())
				assert.NotNil(t, customer.User)
			}
		})
	}
}

func TestCustomerService_List(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.CustomerRepository)
		expectedCount int
		expectedError error
	}{
		{
			name: "Success - List Multiple Customers",
			setupMocks: func(cr *mocks.CustomerRepository) {
				customers := []domain.Customer{
					{
						ID:     uuid.New(),
						UserID: createTestUser().ID,
						User:   createTestUser(),
					},
					{
						ID:     uuid.New(),
						UserID: createTestUser().ID,
						User:   createTestUser(),
					},
				}
				cr.On("List", mock.Anything).Return(customers, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "Success - Empty List",
			setupMocks: func(cr *mocks.CustomerRepository) {
				cr.On("List", mock.Anything).Return([]domain.Customer{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, customerRepo, _ := setupCustomerTest(t)
			tt.setupMocks(customerRepo)

			customers, err := service.List(context.Background())

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, customers)
			} else {
				assert.NoError(t, err)
				assert.Len(t, customers, tt.expectedCount)
				if tt.expectedCount > 0 {
					for _, c := range customers {
						assert.NotNil(t, c.User)
					}
				}
			}
		})
	}
}

func TestCustomerService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		customerID    string
		setupMocks    func(*mocks.CustomerRepository, string)
		expectedError error
	}{
		{
			name:       "Success - Delete Customer",
			customerID: uuid.New().String(),
			setupMocks: func(cr *mocks.CustomerRepository, customerID string) {
				customer := &domain.Customer{
					ID:     uuid.MustParse(customerID),
					UserID: createTestUser().ID,
				}
				cr.On("GetByID", mock.Anything, customerID).
					Return(customer, nil)
				cr.On("Delete", mock.Anything, customerID).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Error - Empty CustomerID",
			customerID:    "",
			setupMocks:    func(_ *mocks.CustomerRepository, _ string) {},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name:       "Error - Customer Not Found",
			customerID: uuid.New().String(),
			setupMocks: func(cr *mocks.CustomerRepository, customerID string) {
				cr.On("GetByID", mock.Anything, customerID).
					Return(nil, customErrors.ErrCustomerNotFound)
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, customerRepo, _ := setupCustomerTest(t)
			tt.setupMocks(customerRepo, tt.customerID)

			err := service.Delete(context.Background(), tt.customerID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
