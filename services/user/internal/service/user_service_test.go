package service

import (
	"errors"
	"testing"
	"time"

	"github.com/bekbull/online-shop/services/user/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of domain.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(page, pageSize int, emailFilter string) ([]*domain.User, int, error) {
	args := m.Called(page, pageSize, emailFilter)
	return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
}

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Test case: Successful user creation
	t.Run("Successful creation", func(t *testing.T) {
		mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("not found"))
		mockRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

		user, err := userService.CreateUser("test@example.com", "Test", "User", "password123", []string{"user"})

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test", user.FirstName)
		assert.Equal(t, "User", user.LastName)
		assert.NotEmpty(t, user.PasswordHash)
		assert.Equal(t, []string{"user"}, user.Roles)
		mockRepo.AssertExpectations(t)
	})

	// Test case: User with email already exists
	t.Run("Email already exists", func(t *testing.T) {
		existingUser := &domain.User{
			ID:        "1",
			Email:     "existing@example.com",
			FirstName: "Existing",
			LastName:  "User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockRepo.On("GetByEmail", "existing@example.com").Return(existingUser, nil)

		user, err := userService.CreateUser("existing@example.com", "Test", "User", "password123", []string{"user"})

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	// Test case: Invalid input - empty email
	t.Run("Empty email", func(t *testing.T) {
		user, err := userService.CreateUser("", "Test", "User", "password123", []string{"user"})

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "email is required")
	})

	// Test case: Invalid input - empty first name
	t.Run("Empty first name", func(t *testing.T) {
		user, err := userService.CreateUser("test@example.com", "", "User", "password123", []string{"user"})

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "first name is required")
	})

	// Test case: Invalid input - empty last name
	t.Run("Empty last name", func(t *testing.T) {
		user, err := userService.CreateUser("test@example.com", "Test", "", "password123", []string{"user"})

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "last name is required")
	})

	// Test case: Invalid input - empty password
	t.Run("Empty password", func(t *testing.T) {
		user, err := userService.CreateUser("test@example.com", "Test", "User", "", []string{"user"})

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "password is required")
	})

	// Test case: Invalid input - password too short
	t.Run("Password too short", func(t *testing.T) {
		user, err := userService.CreateUser("test@example.com", "Test", "User", "123", []string{"user"})

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "password must be at least 8 characters")
	})
}

func TestGetUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Test case: Successful user retrieval
	t.Run("Successful retrieval", func(t *testing.T) {
		mockUser := &domain.User{
			ID:        "user-id-123",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockRepo.On("GetByID", "user-id-123").Return(mockUser, nil)

		user, err := userService.GetUser("user-id-123")

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user-id-123", user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		mockRepo.AssertExpectations(t)
	})

	// Test case: User not found
	t.Run("User not found", func(t *testing.T) {
		mockRepo.On("GetByID", "non-existent").Return(nil, errors.New("user not found"))

		user, err := userService.GetUser("non-existent")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to get user")
		mockRepo.AssertExpectations(t)
	})

	// Test case: Empty ID
	t.Run("Empty ID", func(t *testing.T) {
		user, err := userService.GetUser("")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user ID is required")
	})
}

func TestVerifyPassword(t *testing.T) {
	userService := NewUserService(nil) // Repository not needed for this test

	// Create a user with a known password hash
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           "user-id",
		PasswordHash: string(passwordHash),
	}

	// Test case: Correct password
	t.Run("Correct password", func(t *testing.T) {
		result := userService.VerifyPassword(user, "correct-password")
		assert.True(t, result)
	})

	// Test case: Incorrect password
	t.Run("Incorrect password", func(t *testing.T) {
		result := userService.VerifyPassword(user, "wrong-password")
		assert.False(t, result)
	})
}
