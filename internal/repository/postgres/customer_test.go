package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/grocery-service/internal/domain"
	"github.com/grocery-service/internal/repository/db"
	customErrors "github.com/grocery-service/utils/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupCustomerTestDB(t *testing.T) *db.PostgresDB {
	postgres, err := db.NewTestDB()
	require.NoError(t, err)

	err = postgres.DB.Migrator().DropTable(&domain.Customer{}, &domain.User{})
	require.NoError(t, err)
	err = postgres.DB.AutoMigrate(&domain.User{}, &domain.Customer{})
	require.NoError(t, err)

	return postgres
}

func createTestUser(t *testing.T, db *gorm.DB) *domain.User {
	user := &domain.User{
		ID:      uuid.New(),
		Email:   "test@example.com",
		Name:    "Test User",
		Role:    domain.CustomerRole,
		Picture: "https://example.com/picture.jpg",
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

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
			setupTest: func(db *gorm.DB) *domain.Customer {
				return &domain.Customer{
					ID:     uuid.New(),
					UserID: uuid.New(), // Non-existent user ID
				}
			},
			expectedError: customErrors.ErrInvalidCustomerData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupCustomerTestDB(t)
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
			setupTest: func(db *gorm.DB) (string, *domain.Customer) {
				return uuid.New().String(), nil
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupCustomerTestDB(t)
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
			setupTest: func(db *gorm.DB) (string, *domain.Customer) {
				return uuid.New().String(), nil
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupCustomerTestDB(t)
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
					{ID: uuid.New(), UserID: createTestUser(t, db).ID},
					{ID: uuid.New(), UserID: createTestUser(t, db).ID},
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
			setupTest: func(db *gorm.DB) []domain.Customer {
				return []domain.Customer{}
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupCustomerTestDB(t)
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
			setupTest: func(db *gorm.DB) *domain.Customer {
				return &domain.Customer{ID: uuid.New()}
			},
			updateFunc:    func(db *gorm.DB, c *domain.Customer) {},
			expectedError: customErrors.ErrCustomerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupCustomerTestDB(t)
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
			setupTest: func(db *gorm.DB) string {
				return uuid.New().String()
			},
			expectedError: customErrors.ErrCustomerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgres := setupCustomerTestDB(t)
			repo := NewCustomerRepository(postgres)
			ctx := context.Background()

			id := tt.setupTest(postgres.DB)
			err := repo.Delete(ctx, id)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				err = postgres.DB.First(&domain.Customer{}, "id = ?", id).Error
				assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
			}
		})
	}
}
