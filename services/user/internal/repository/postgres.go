package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bekbull/online-shop/services/user/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// PostgresRepository implements the UserRepository interface using PostgreSQL
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// Create inserts a new user into the database
func (r *PostgresRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, email, first_name, last_name, password_hash, roles, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.PasswordHash,
		pq.Array(user.Roles),
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate email
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email") {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *PostgresRepository) GetByID(id string) (*domain.User, error) {
	query := `
		SELECT id, email, first_name, last_name, password_hash, roles, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	var roles []byte // Store the roles as a byte array initially

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&roles, // Roles will be parsed separately
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Parse the PostgreSQL array
	var roleArray pq.StringArray
	err = roleArray.Scan(roles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse roles: %w", err)
	}

	// Convert to string slice
	user.Roles = []string(roleArray)

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *PostgresRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, email, first_name, last_name, password_hash, roles, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	var roles []byte // Store the roles as a byte array initially

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&roles, // Roles will be parsed separately
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Parse the PostgreSQL array
	var roleArray pq.StringArray
	err = roleArray.Scan(roles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse roles: %w", err)
	}

	// Convert to string slice
	user.Roles = []string(roleArray)

	return &user, nil
}

// Update updates a user in the database
func (r *PostgresRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, first_name = $3, last_name = $4, password_hash = $5, roles = $6, updated_at = $7
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.PasswordHash,
		pq.Array(user.Roles),
		user.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate email
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "email") {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete removes a user from the database
func (r *PostgresRepository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", id)
	}

	return nil
}

// List retrieves a list of users with pagination and optional filtering
func (r *PostgresRepository) List(page, pageSize int, emailFilter string) ([]*domain.User, int, error) {
	// Ensure valid pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Base query
	query := `
		SELECT id, email, first_name, last_name, password_hash, roles, created_at, updated_at
		FROM users
	`
	countQuery := `SELECT COUNT(*) FROM users`

	// Add filter if provided
	var args []interface{}
	if emailFilter != "" {
		query += ` WHERE email ILIKE $1`
		countQuery += ` WHERE email ILIKE $1`
		args = append(args, "%"+emailFilter+"%")
	}

	// Add pagination
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, pageSize, offset)

	// Get total count
	var totalCount int
	err := r.db.Get(&totalCount, countQuery, args[:max(0, len(args)-2)]...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	// Process results
	var users []*domain.User
	for rows.Next() {
		var user domain.User
		var roles []byte // Store the roles as a byte array initially

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.PasswordHash,
			&roles, // Roles will be parsed separately
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		// Parse the PostgreSQL array
		var roleArray pq.StringArray
		err = roleArray.Scan(roles)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to parse roles: %w", err)
		}

		// Convert to string slice
		user.Roles = []string(roleArray)

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, totalCount, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// InitDB initializes the database schema
func (r *PostgresRepository) InitDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		first_name VARCHAR(100) NOT NULL,
		last_name VARCHAR(100) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		roles TEXT[] NOT NULL DEFAULT '{}',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := r.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return nil
}
