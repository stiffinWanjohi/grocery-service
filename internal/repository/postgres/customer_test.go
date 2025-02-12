package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCustomerRepository_Create(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) *domain.Customer
		expectedError error
	}{
		{
			name: "Success - Create Customer",
			setupTest: func(db *gorm.DB) *domain.Customer {
				user := createTestUser(t, db)
				return &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
				}
			},
			expectedError: nil,
		},
		{
			name: "Error - Invalid User ID",
			setupTest: func(_ *gorm.DB) *domain.Customer {
				return &domain.Customer{
					ID:     uuid.New(),
					UserID: uuid.New(),
				}
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name: "Error - Empty User ID",
			setupTest: func(_ *gorm.DB) *domain.Customer {
				return &domain.Customer{
					ID:     uuid.New(),
					UserID: uuid.UUID{},
				}
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name: "Error - Empty Customer ID",
			setupTest: func(db *gorm.DB) *domain.Customer {
				user := createTestUser(t, db)
				return &domain.Customer{
					ID:     uuid.UUID{},
					UserID: user.ID,
				}
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Customer{}, &domain.User{})
			repo := NewCustomerRepository(postgres)
			ctx := context.Background()

			customer := tt.setupTest(postgres.DB)
			err := repo.Create(ctx, customer)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)

				var found domain.Customer
				err = postgres.DB.Preload("User").First(&found, "id = ?", customer.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, customer.UserID, found.UserID)
				assert.NotNil(t, found.User)
			}
		})
	}
}

func TestCustomerRepository_GetByID(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (string, *domain.Customer)
		expectedError error
	}{
		{
			name: "Success - Get Existing Customer",
			setupTest: func(db *gorm.DB) (string, *domain.Customer) {
				user := createTestUser(t, db)
				customer := &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
				}
				require.NoError(t, db.Create(customer).Error)
				return customer.ID.String(), customer
			},
			expectedError: nil,
		},
		{
			name: "Error - Customer Not Found",
			setupTest: func(_ *gorm.DB) (string, *domain.Customer) {
				return uuid.New().String(), nil
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
		{
			name: "Error - Invalid UUID",
			setupTest: func(_ *gorm.DB) (string, *domain.Customer) {
				return "invalid-uuid", nil
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Customer{}, &domain.User{})
			repo := NewCustomerRepository(postgres)
			ctx := context.Background()

			id, expected := tt.setupTest(postgres.DB)
			found, err := repo.GetByID(ctx, id)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, expected.ID, found.ID)
				assert.NotNil(t, found.User)
			}
		})
	}
}

func TestCustomerRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) (string, *domain.Customer)
		expectedError error
	}{
		{
			name: "Success - Get Customer By UserID",
			setupTest: func(db *gorm.DB) (string, *domain.Customer) {
				user := createTestUser(t, db)
				customer := &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
				}
				require.NoError(t, db.Create(customer).Error)
				return user.ID.String(), customer
			},
			expectedError: nil,
		},
		{
			name: "Error - No Customer For UserID",
			setupTest: func(_ *gorm.DB) (string, *domain.Customer) {
				return uuid.New().String(), nil
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
		{
			name: "Error - Invalid UUID Format",
			setupTest: func(_ *gorm.DB) (string, *domain.Customer) {
				return "invalid-uuid", nil
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Customer{}, &domain.User{})
			repo := NewCustomerRepository(postgres)
			ctx := context.Background()

			userID, expected := tt.setupTest(postgres.DB)
			found, err := repo.GetByUserID(ctx, userID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, expected.ID, found.ID)
				assert.NotNil(t, found.User)
			}
		})
	}
}

func TestCustomerRepository_List(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) []domain.Customer
		expectedCount int
	}{
		{
			name: "Success - List Multiple Customers",
			setupTest: func(db *gorm.DB) []domain.Customer {
				customers := []domain.Customer{
					{
						ID:     uuid.New(),
						UserID: createTestUser(t, db).ID,
					},
					{
						ID:     uuid.New(),
						UserID: createTestUser(t, db).ID,
					},
				}
				for _, c := range customers {
					require.NoError(t, db.Create(&c).Error)
				}
				return customers
			},
			expectedCount: 2,
		},
		{
			name: "Success - Empty List",
			setupTest: func(_ *gorm.DB) []domain.Customer {
				return []domain.Customer{}
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Customer{}, &domain.User{})
			repo := NewCustomerRepository(postgres)
			ctx := context.Background()

			_ = tt.setupTest(postgres.DB)
			found, err := repo.List(ctx)

			assert.NoError(t, err)
			assert.Len(t, found, tt.expectedCount)
			if tt.expectedCount > 0 {
				for _, customer := range found {
					assert.NotNil(t, customer.User)
				}
			}
		})
	}
}

func TestCustomerRepository_Update(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) *domain.Customer
		updateFunc    func(*gorm.DB, *domain.Customer)
		expectedError error
	}{
		{
			name: "Success - Update Customer",
			setupTest: func(db *gorm.DB) *domain.Customer {
				user := createTestUser(t, db)
				customer := &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
				}
				require.NoError(t, db.Create(customer).Error)
				return customer
			},
			updateFunc: func(db *gorm.DB, c *domain.Customer) {
				c.UserID = createTestUser(t, db).ID
			},
			expectedError: nil,
		},
		{
			name: "Error - Customer Not Found",
			setupTest: func(_ *gorm.DB) *domain.Customer {
				return &domain.Customer{
					ID:     uuid.New(),
					UserID: uuid.New(),
				}
			},
			updateFunc: func(db *gorm.DB, c *domain.Customer) {
				user := createTestUser(t, db)
				c.UserID = user.ID
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
		{
			name: "Error - Invalid UUID",
			setupTest: func(_ *gorm.DB) *domain.Customer {
				return &domain.Customer{
					ID:     uuid.New(),
					UserID: uuid.New(),
				}
			},
			updateFunc: func(_ *gorm.DB, c *domain.Customer) {
				c.ID = uuid.UUID{} // Invalid UUID
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name: "Error - Invalid User ID",
			setupTest: func(db *gorm.DB) *domain.Customer {
				user := createTestUser(t, db)
				customer := &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
				}
				require.NoError(t, db.Create(customer).Error)
				return customer
			},
			updateFunc: func(_ *gorm.DB, c *domain.Customer) {
				c.UserID = uuid.UUID{} // Invalid User ID
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
		{
			name: "Error - Non-Existent User ID",
			setupTest: func(db *gorm.DB) *domain.Customer {
				user := createTestUser(t, db)
				customer := &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
				}
				require.NoError(t, db.Create(customer).Error)
				return customer
			},
			updateFunc: func(_ *gorm.DB, c *domain.Customer) {
				c.UserID = uuid.New() // Valid UUID format but non-existent user
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Customer{}, &domain.User{})
			repo := NewCustomerRepository(postgres)
			ctx := context.Background()

			customer := tt.setupTest(postgres.DB)
			tt.updateFunc(postgres.DB, customer)

			err := repo.Update(ctx, customer)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				var found domain.Customer
				err = postgres.DB.Preload("User").First(&found, "id = ?", customer.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, customer.UserID, found.UserID)
			}
		})
	}
}

func TestCustomerRepository_Delete(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*gorm.DB) string
		expectedError error
	}{
		{
			name: "Success - Delete Customer",
			setupTest: func(db *gorm.DB) string {
				user := createTestUser(t, db)
				customer := &domain.Customer{
					ID:     uuid.New(),
					UserID: user.ID,
				}
				require.NoError(t, db.Create(customer).Error)
				return customer.ID.String()
			},
			expectedError: nil,
		},
		{
			name: "Error - Customer Not Found",
			setupTest: func(_ *gorm.DB) string {
				return uuid.New().String()
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
		{
			name: "Error - Invalid UUID",
			setupTest: func(_ *gorm.DB) string {
				return "invalid-uuid"
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupTestDB(t, &domain.Customer{}, &domain.User{})
			repo := NewCustomerRepository(postgres)
			ctx := context.Background()

			id := tt.setupTest(postgres.DB)
			err := repo.Delete(ctx, id)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				// Skip database check for invalid UUID case
				if tt.name != "Error - Invalid UUID" {
					err = postgres.DB.First(
						&domain.Customer{},
						"id = ?",
						id,
					).Error
					assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
				}
			} else {
				assert.NoError(t, err)
				err = postgres.DB.First(&domain.Customer{}, "id = ?", id).Error
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}
