package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Roles        []string  `json:"roles" db:"roles"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// NewUser creates a new user with default values
func NewUser(email, firstName, lastName, passwordHash string, roles []string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: passwordHash,
		Roles:        roles,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	List(page, pageSize int, emailFilter string) ([]*User, int, error)
}

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(email, firstName, lastName, password string, roles []string) (*User, error)
	GetUser(id string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	UpdateUser(id string, updates map[string]interface{}) (*User, error)
	DeleteUser(id string) error
	ListUsers(page, pageSize int, emailFilter string) ([]*User, int, error)
}
