package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/bekbull/online-shop/services/user/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// UserService implements the UserService interface
type UserService struct {
	repo domain.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(email, firstName, lastName, password string, roles []string) (*domain.User, error) {
	// Validate input
	if email == "" {
		return nil, errors.New("email is required")
	}
	if firstName == "" {
		return nil, errors.New("first name is required")
	}
	if lastName == "" {
		return nil, errors.New("last name is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := domain.NewUser(email, firstName, lastName, string(hashedPassword), roles)

	// Save to repository
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(id string) (*domain.User, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*domain.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// UpdateUser updates a user's details
func (s *UserService) UpdateUser(id string, updates map[string]interface{}) (*domain.User, error) {
	if id == "" {
		return nil, errors.New("user ID is required")
	}

	// Get existing user
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "email":
			if email, ok := value.(string); ok && email != "" {
				// Check if email is already taken by another user
				existingUser, err := s.repo.GetByEmail(email)
				if err == nil && existingUser != nil && existingUser.ID != id {
					return nil, fmt.Errorf("email %s is already taken", email)
				}
				user.Email = email
			}
		case "first_name":
			if firstName, ok := value.(string); ok && firstName != "" {
				user.FirstName = firstName
			}
		case "last_name":
			if lastName, ok := value.(string); ok && lastName != "" {
				user.LastName = lastName
			}
		case "password":
			if password, ok := value.(string); ok && password != "" {
				if len(password) < 8 {
					return nil, errors.New("password must be at least 8 characters")
				}
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					return nil, fmt.Errorf("failed to hash password: %w", err)
				}
				user.PasswordHash = string(hashedPassword)
			}
		case "roles":
			if roles, ok := value.([]string); ok {
				user.Roles = roles
			}
		}
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save to repository
	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id string) error {
	if id == "" {
		return errors.New("user ID is required")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a list of users with pagination and optional filtering
func (s *UserService) ListUsers(page, pageSize int, emailFilter string) ([]*domain.User, int, error) {
	users, total, err := s.repo.List(page, pageSize, emailFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// VerifyPassword checks if the provided password matches the stored hash
func (s *UserService) VerifyPassword(user *domain.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}
